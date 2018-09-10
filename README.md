# Live Queries

This package attempts to add a basic version of live queries to Ponzu. By 
interacting with the database hooks provided with the `item.Hookable` interface, 
the `live` package (`import "github.com/ponzu-cms/live"`) enables users to publish
and subscribe to the hookable events throughout the content package.

Usage:

```go
package content

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ponzu-cms/live"
	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/item"
)

// typical generated ponzu content type
type Customer struct {
	item.Item

	Name  string `json:"name"`
	Email string `json:"email"`
}

// declare a package-level subscriptions container (could be in it's own file 
// within the content package)
var subs live.Subscriptions

func init() {
    item.Types["Customer"] = func() interface{} { return new(Customer) }

    // create the new subscriptions container to Subscribe / Emit
    subs = live.New()

    // create a subscriber to handle AdminUpdate events on a Customer
    go func() {
        adminUpdates := subs.Subscribe("Customer", live.AdminUpdate)
        for event := range adminUpdates {
            log.Println("<g1> UPDATE:", event.Content().(*Customer))
        }
    }()

    // NOTE: any number of subscriptions can be made to the same content type for
    // any event type.

    // create another subscriber for AdminDelete events on a Customer
    go func() {
        adminDeletes := subs.Subscribe("Customer", live.AdminDelete)
        for event := range adminDeletes {
            log.Println("<g2> DELETE:", event.Content().(*Customer))
        }
    }()
}

// emit content events from within the item.Hookable methods provided by Ponzu 
// to broadcast a change (hookable constants are defined in `live` for event types)
func (c *Customer) AfterAdminUpdate(res http.ResponseWriter, req *http.Request) error {
	err := subs.Emit(req.Context(), "Customer", c, live.AdminUpdate)
	if err != nil {
		if liveErr, ok := err.(live.QueryError); ok {
			// handle live error in specific manner
			fmt.Println(liveErr)
			return nil
		}

		return err
	}

	return nil
}

func (c *Customer) BeforeAdminDelete(res http.ResponseWriter, req *http.Request) error {
	err := subs.Emit(req.Context(), "Customer", c, live.AdminDelete)
	if err != nil {
		return err
	}

	return nil
}

// MarshalEditor writes a buffer of html to edit a Customer within the CMS
// and implements editor.Editable
func (c *Customer) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Customer field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", c, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.Input("Email", c, map[string]string{
				"label":       "Email",
				"type":        "text",
				"placeholder": "Enter the Email here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Customer editor view: %s", err.Error())
	}

	return view, nil
}

// String defines how a Customer is printed. Update it using more descriptive
// fields from the Customer struct type
func (c *Customer) String() string {
	return fmt.Sprintf("Customer: %s", c.UUID)
}
```