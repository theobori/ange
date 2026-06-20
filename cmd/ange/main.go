package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/theobori/ange/internal/hash"
	"github.com/theobori/ange/internal/random"
	"github.com/theobori/ange/internal/webring"
	fleurform "github.com/theobori/fleur-form"
	"github.com/theobori/fleur/gopher"
	gserver "github.com/theobori/fleur/gopher/server"
	"github.com/theobori/fleur/gophermap"
	"github.com/theobori/fleur/gophermap/evaluator"
	"github.com/theobori/fleur/server"
)

func createWebringSubmitCallback(w *webring.Webring) fleurform.SubmitCallback {
	return func(parameters fleurform.Parameters, server *server.Server, ctx *server.RequestContext) error {
		port, err := strconv.Atoi(parameters["port"])
		if err != nil {
			return err
		}

		path := "/" + strings.Trim(parameters["path"], "/")

		gopherspace, err := webring.NewGopherspaceFromUserInput(parameters["title"], parameters["domain"], port, path)
		if err != nil {
			return server.SendError(ctx.Conn, err.Error())
		}

		id, err := w.AddForReview(gopherspace)
		if err != nil {
			return server.SendError(ctx.Conn, err.Error())
		}

		menu := gophermap.RenderMenu(
			server.NewItem(gophermap.ItemTypeInlineText, "You successfully submitted your gopherspace for a review.", "/"),
			server.NewItem(gophermap.ItemTypeInlineText, "", "/"),
			server.NewItem(gophermap.ItemTypeInlineText, "Make sure to remember your gopherspace id and token below:", "/"),
			server.NewItem(gophermap.ItemTypeInlineText, fmt.Sprintf("Your ID: %d", id), "/"),
			server.NewItem(gophermap.ItemTypeInlineText, fmt.Sprintf("Your token: %s", gopherspace.Token), "/"),
		)

		return gserver.SendString(ctx.Conn, menu)
	}
}

func createAdminPanelSubmitCallback(w *webring.Webring) fleurform.SubmitCallback {
	return func(parameters fleurform.Parameters, server *server.Server, ctx *server.RequestContext) error {
		if !w.IsAdmin(parameters["secret"]) {
			return server.SendError(ctx.Conn, "You must be admin.")
		}

		id, err := strconv.Atoi(parameters["id"])
		if err != nil {
			return err
		}

		switch parameters["action"] {
		case "deny":
			err := w.Deny(id)
			if err != nil {
				return err
			}
			server.SendGophermap(ctx.Conn, gophermap.ItemTypeInlineText, fmt.Sprintf("You successfully denied the gopherspace with id '%d'.", id))
		case "approve":
			err := w.Approve(id)
			if err != nil {
				return err
			}
			server.SendGophermap(ctx.Conn, gophermap.ItemTypeInlineText, fmt.Sprintf("You successfully approved the gopherspace with id '%d'.", id))
		default:
			return fmt.Errorf("Invalid action value.")
		}

		return nil
	}
}

func createNeighborRouteCallback(w *webring.Webring, isNext bool) server.RouteCallback {
	var name string

	if isNext {
		name = "next"
	} else {
		name = "previous"
	}

	return func(server *server.Server, ctx *server.RequestContext) error {
		idString := strings.TrimPrefix(ctx.VirtualPath, "/webring/"+name+"/")
		id, err := strconv.Atoi(idString)
		if err != nil {
			return err
		}

		gopherspace, err := w.NeighborGopherspace(id, isNext)
		if err != nil {
			return err
		}

		return gserver.SendString(
			ctx.Conn,
			gopherspace.GophermapItem().String(),
		)
	}
}

func renderGopherspacesList(w *webring.Webring, approved bool) (string, error) {
	items := []*gophermap.Item{}

	gopherspaces, err := w.ListGopherspaces(approved)
	if err != nil {
		return "", err
	}

	for _, gopherspace := range gopherspaces {
		item := gophermap.Item{
			ItemType:    gophermap.ItemTypeGopherMenu,
			Description: fmt.Sprintf("%04d %s", gopherspace.Id, gopherspace.Title),
			Selector:    gopherspace.Path,
			Domain:      gopherspace.Domain,
			Port:        gopherspace.Port,
		}
		items = append(items, &item)
	}

	if len(items) == 0 {
		var description string
		if approved {
			description = "There are no gopherspaces in the webring."
		} else {
			description = "There are no gopherspaces that need review."
		}

		items = append(items, &gophermap.Item{
			ItemType:    gophermap.ItemTypeInlineText,
			Description: description,
			Selector:    "/",
			Domain:      "/",
			Port:        0,
		})
	}

	return gophermap.RenderMenu(items...), nil
}

func getSecret() (*hash.Hash, error) {
	secretString, ok := os.LookupEnv("ANGE_SECRET")
	if !ok {
		secretString = random.Generate(32)
		log.Println("It's recommended to set your own secret by setting the ANGE_SECRET environement variable.")
		log.Println("A secret has been automatically generated for you:")
		log.Println(secretString)
	}

	if len(secretString) < 10 {
		return nil, fmt.Errorf("Your secret should have at least a size of 10.")
	}

	secret := hash.NewHash(secretString)

	return secret, nil
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

	secret, err := getSecret()
	if err != nil {
		log.Fatalln(err)
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
			return renderGopherspacesList(w, true)
		},
	)

	router := server.NewRouter()
	err = fleurform.Apply(
		router,
		"/webring",
		[]fleurform.ParameterMetadata{
			{
				Name:     "title",
				Required: true,
			},
			{
				Name:     "domain",
				Required: true,
			},
			{
				Name:     "port",
				Required: true,
			},
			{
				Name:     "path",
				Required: true,
			},
		},
		createWebringSubmitCallback(w),
	)
	if err != nil {
		log.Fatalln(err)
	}
	err = fleurform.Apply(
		router,
		"/webring/adminpanel",
		[]fleurform.ParameterMetadata{
			{
				Name:     "secret",
				Required: true,
			},
			{
				Name:     "id",
				Required: true,
			},
			{
				Name:     "action",
				Required: true,
			},
		},
		createAdminPanelSubmitCallback(w),
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

	router.SetWithWeight(
		1,
		"^/webring/adminpanel$",
		func(server *server.Server, ctx *server.RequestContext) error {
			if len(ctx.SearchParameter) == 0 {
				return server.SendError(ctx.Conn, "Missing search parameter.")
			}

			if !w.IsAdmin(ctx.SearchParameter) {
				return server.SendError(ctx.Conn, "You must be admin.")
			}

			entryList, err := renderGopherspacesList(w, false)
			if err != nil {
				return err
			}

			menu := gophermap.RenderMenu(
				server.NewItem(gophermap.ItemTypeGopherMenu, "Access the form to review a gopherspace candidate.", "/webring/adminpanel/form"),
				server.NewItem(gophermap.ItemTypeInlineText, "For the action field, you can enter 'deny' or 'approve'.", "/"),
				server.NewItem(gophermap.ItemTypeInlineText, "", "/"),
			) + "\n" + entryList

			return gserver.SendString(ctx.Conn, menu)
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
				return server.SendError(ctx.Conn, "An error occured. Are you sure it's your token?")
			}

			return server.SendGophermap(
				ctx.Conn,
				gophermap.ItemTypeInlineText,
				"You successfully deleted your gopherspace.",
			)
		},
	)
	router.SetWithWeight(
		1,
		"^/webring/next/.*$",
		createNeighborRouteCallback(w, true),
	)
	router.SetWithWeight(
		1,
		"^/webring/previous/.*$",
		createNeighborRouteCallback(w, false),
	)
	router.SetWithWeight(
		1,
		"^/webring/random$",
		func(server *server.Server, ctx *server.RequestContext) error {
			gopherspace, err := w.RandomGopherspace()
			if err != nil {
				return err
			}

			return gserver.SendString(
				ctx.Conn,
				gopherspace.GophermapItem().String(),
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
