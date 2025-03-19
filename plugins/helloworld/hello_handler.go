package helloworld

import (
	"fmt"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
)

type HelloworldPageData struct {
	FormValues HelloworldFormValues
	FormErrors v.Errors
}
type HelloworldFormValues struct {
	Message        string `form:"message"`
	successMessage string
}

var helloworldSchema = v.Schema{
	"message": v.Rules(v.Min(1), v.Max(50)),
}

func handleHello(kit *kit.Kit) error {
	return kit.Render(Helloworld(HelloworldPageData{}))
}
func handlePostHello(kit *kit.Kit) error {
	var values HelloworldFormValues
	errors, ok := v.Request(kit.Request, &values, helloworldSchema)
	if !ok {
		return kit.Render(PostMessage(values, errors))
	}
	_, err := createMessage(values.Message)
	if err != nil {
		return kit.Render(PostMessage(values, errors))
	}
	return kit.Render(PostMessage(HelloworldFormValues{successMessage: fmt.Sprintf("New message created %s", values.Message)}, errors))
}

func handleAuthenticatedHello(kit *kit.Kit) error {
	return kit.Render(HelloworldAuth())
}
func handleReadHello(kit *kit.Kit) error {
	messages, err := listMessages()
	if err != nil {
		panic(err)
	}
	return kit.Render(HelloworldRead(messages))
}
