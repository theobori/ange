package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/theobori/ange/internal/common"
	"github.com/theobori/ange/internal/webring"
	"github.com/theobori/fleur-form"
	"github.com/theobori/fleur/gopher"
	gserver "github.com/theobori/fleur/gopher/server"
	"github.com/theobori/fleur/gophermap"
	"github.com/theobori/fleur/gophermap/evaluator"
	"github.com/theobori/fleur/server"
)

func createSubmitCallback(w *webring.Webring) fleurform.SubmitCallback {
	return func(parameters fleurform.Parameters, server *server.Server, ctx *server.RequestContext) error {
		for _, parameterName := range webring.GopherspaceEntryKeywords {
			_, ok := parameters[parameterName]
			if !ok {
				return server.SendError(
					ctx.Conn,
					fmt.Sprintf("Missing the '%s' parameter.", parameterName),
				)
			}
		}

		port, err := strconv.Atoi(parameters["port"])
		if err != nil {
			return err
		}

		entry := webring.GopherspaceEntry{
			Title:  parameters["title"],
			Domain: parameters["domain"],
			Port:   port,
			Path:   parameters["path"],
		}

		node, err := w.Add(&entry)
		if err != nil {
			fmt.Println(err)
			return server.SendError(ctx.Conn, err.Error())
		}

		menu := gophermap.RenderMenu(
			server.NewItem(gophermap.ItemTypeInlineText, "You successfully submitted your gopherspace for a review.", "/"),
			server.NewItem(gophermap.ItemTypeInlineText, "", "/"),
			server.NewItem(gophermap.ItemTypeInlineText, "Make sure to remember your gopherspace id and token below:", "/"),
			server.NewItem(gophermap.ItemTypeInlineText, fmt.Sprintf("Your ID: %d", node.Id), "/"),
			server.NewItem(gophermap.ItemTypeInlineText, fmt.Sprintf("Your token: %s", node.Token), "/"),
		)

		return gserver.SendString(ctx.Conn, menu)
	}
}

func RenderEntryList(w *webring.Webring) (string, error) {
	items := []*gophermap.Item{}

	nodes, err := w.List()
	if err != nil {
		return "", err
	}

	for _, node := range nodes {
		item := gophermap.Item{
			ItemType:    gophermap.ItemTypeGopherMenu,
			Description: fmt.Sprintf("%04d %s", node.Id, node.Entry.Title),
			Selector:    node.Entry.Path,
			Domain:      node.Entry.Domain,
			Port:        node.Entry.Port,
		}
		items = append(items, &item)
	}

	if len(items) == 0 {
		items = append(items, &gophermap.Item{
			ItemType:    gophermap.ItemTypeInlineText,
			Description: "The webring is empty",
			Selector:    "/",
			Domain:      "/",
			Port:        0,
		})
	}

	return gophermap.RenderMenu(items...), nil
}

func main() {
	var (
		err           error
		domain        string
		directoryPath string
		databasePath  string
		port          int
		verbose       bool
	)

	flag.StringVar(
		&domain,
		"domain",
		"localhost",
		"Gopher domain.",
	)
	flag.StringVar(
		&directoryPath,
		"directory",
		"./",
		"It specifies an input directory path that will be the root of the Gopher server.",
	)
	flag.StringVar(
		&databasePath,
		"database-path",
		"./ange-database.db",
		"SQLite3 database path.",
	)
	flag.IntVar(
		&port,
		"port",
		gopher.DefaultPort,
		"Gopher port.",
	)
	flag.BoolVar(
		&verbose,
		"verbose",
		false,
		"Enable verbose logs.",
	)

	flag.Parse()

	if port < 0 {
		log.Fatalln("The port should at least be a positive integer.")
	}

	secret, ok := os.LookupEnv("ANGE_SECRET")
	if !ok {
		secret = common.GenerateToken(32)
		log.Println("It's recommended to set your own secret by setting the ANGE_SECRET environement variable.")
		log.Println("A secret has been automatically generated for you:")
		log.Println(secret)
	}

	if len(secret) < 10 {
		log.Fatalln("Your secret should have at least a size of 10.")
	}

	serverOptions, err := server.NewOptions(
		port,
		directoryPath,
		domain,
		verbose,
	)
	if err != nil {
		log.Fatalln(err)
	}

	evaluatorOptions := evaluator.Options{
		Port:                 serverOptions.Port,
		DirectoryPath:        serverOptions.DirectoryPath,
		Domain:               serverOptions.Domain,
		EnableAutoInlineText: true,
	}

	database, err := webring.NewDatabase(databasePath)
	if err != nil {
		log.Fatalln(err)
	}

	w, err := webring.NewWebring(database, secret)
	if err != nil {
		log.Fatalln(err)
	}

	em := evaluator.RFC1436ItemsExtensionManager()
	em.Set(
		"^>members",
		func(e *evaluator.Evaluator, ctx *evaluator.ExtensionContext) (string, error) {
			return RenderEntryList(w)
		},
	)

	router := server.NewRouter()
	err = fleurform.AddFormToRouter(
		router,
		"/webring",
		webring.GopherspaceEntryKeywords,
		createSubmitCallback(w),
	)
	if err != nil {
		log.Fatalln(err)
	}

	router.SetWithWeight(
		0,
		".*",
		func(server *server.Server, ctx *server.RequestContext) error {
			ctx.Path = filepath.Join(server.Options().DirectoryPath, "gophermap")
			ctx.VirtualPath = "gophermap"

			return server.HandleFile(ctx)
		},
	)

	// TODO: form admin avec id et secret
	router.SetWithWeight(
		1,
		"^/webring/deny$",
		func(server *server.Server, ctx *server.RequestContext) error {
			return nil
		},
	)
	router.SetWithWeight(
		1,
		"^/webring/accept$",
		func(server *server.Server, ctx *server.RequestContext) error {
			return nil
		},
	)
	router.SetWithWeight(
		1,
		"^/webring/review$",
		func(server *server.Server, ctx *server.RequestContext) error {
			if len(ctx.SearchParameter) == 0 {
				return server.SendError(ctx.Conn, "Missing search parameter.")
			}

			return nil
		},
	)
	router.SetWithWeight(
		1,
		"^/webring/delete$",
		func(server *server.Server, ctx *server.RequestContext) error {
			if len(ctx.SearchParameter) == 0 {
				return server.SendError(ctx.Conn, "Missing search parameter.")
			}

			token := ctx.SearchParameter

			err := w.Delete(token)
			if err != nil {
				fmt.Println(err)
				return server.SendError(ctx.Conn, "An error occured. Are you sure it's your token?")
			}

			return server.SendGophermap(
				ctx.Conn,
				gophermap.ItemTypeInlineText,
				fmt.Sprintf("You successfully deleted the gopherspace with token '%s'.", token),
			)
		},
	)

	evaluator := evaluator.NewEvaluator(&evaluatorOptions, em)
	server := server.NewServerWithRouter(serverOptions, evaluator, router)

	err = server.Serve()
	if err != nil {
		log.Fatalln(err)
	}
}
