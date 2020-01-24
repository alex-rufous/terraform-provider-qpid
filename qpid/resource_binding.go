package qpid

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
	"strings"
)

type Binding struct {
	BindingKey      string
	Destination     string
	Exchange        string
	Arguments       map[string]string
	VirtualHostNode string
	VirtualHost     string
}

func resourceBinding() *schema.Resource {
	return &schema.Resource{
		Create: createBinding,
		Read:   readBinding,
		Delete: deleteBinding,
		Update: updateBinding,
		Exists: existsBinding,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"binding_key": {
				Type:        schema.TypeString,
				Description: "Binding key",
				Required:    true,
				ForceNew:    true,
			},

			"destination": {
				Type:        schema.TypeString,
				Description: "Destination",
				Required:    true,
				ForceNew:    true,
			},

			"exchange": {
				Type:        schema.TypeString,
				Description: "Exchange",
				Required:    true,
				ForceNew:    true,
			},

			"arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},

			"parents": {
				Type:        schema.TypeList,
				Description: "Parents of Binding: <node>, <host>",
				Required:    true,
				ForceNew:    true,
				MaxItems:    2,
				MinItems:    2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func createBinding(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	b, err := buildBinding(d)
	if err != nil {
		return err
	}
	resp, err := client.CreateBinding(b)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK{


		defer func() {
			closeError := resp.Body.Close()
			if err == nil {
				err = closeError
			}
		}()
		var result bool
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return err
		}

		id := b.VirtualHostNode +"|" + b.VirtualHost + "|" + b.Exchange + "|" + b.Destination + "|" + b.BindingKey
		d.SetId(id)
		if !result {
			return fmt.Errorf("binding already exist for destination '%s' and exchange '%s' with binding key '%s' on host '%s/%s'",
				b.Destination, b.Exchange, b.BindingKey, b.VirtualHostNode, b.VirtualHost)
		}
		return nil
	}

	return fmt.Errorf("error creating binding of destination '%s' to exchange '%s' with key '%s' on host '%s/%s': %s",
		b.Destination, b.Exchange, b.BindingKey, b.VirtualHostNode, b.VirtualHost, resp.Status)
}

func buildBinding(d *schema.ResourceData) (*Binding, error) {
	bindingKey := d.Get("binding_key").(string)
	destination := d.Get("destination").(string)
	exchange := d.Get("exchange").(string)
	arguments := d.Get("arguments").(map[string]interface{})
	parentItems := d.Get("parents").([]interface{})
	var parents = *convertToArrayOfStrings(&parentItems)

	if len(parents) != 2 {
		return &Binding{}, fmt.Errorf("unexpected exchange parents: %s", strings.Join(parents, "/"))
	}
	args := *convertToMapOfStrings(&arguments)
	return &Binding{
		bindingKey,
		destination,
		  exchange,
		args,
		parents[0],
		parents[1]}, nil
}

func readBinding(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	id := d.Id()
	parts := strings.Split(id, "|")
	if len(parts) != 5 {
		return fmt.Errorf("unexpected id '%s' for qpid binding", id)
	}

	b, err := buildBinding(d)
	if err != nil {
		return err
	}

	b, err = client.GetBinding(b)
	if err != nil {
		return err
	}

	if b == nil {
		return nil
	}

	err = d.Set("binding_key", b.BindingKey)
	if err != nil {
		return err
	}

	err = d.Set("exchange", b.Exchange)
	if err != nil {
		return err
	}

	err = d.Set("destination", b.Destination)
	if err != nil {
		return err
	}

	err = d.Set("arguments", b.Arguments)
	if err != nil {
		return err
	}

	err = d.Set("parents", []interface{}{b.VirtualHostNode, b.VirtualHost})
	if err != nil {
		return err
	}

	return nil
}

func existsBinding(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	id := d.Id()
	parts := strings.Split(id, "|")
	if len(parts) != 5 {
		return false, fmt.Errorf("unexpected id '%s' for qpid binding", id)
	}

	b, err := buildBinding(d)
	if err != nil {
		return false, err
	}
	binding, err := client.GetBinding(b)
	if err != nil {
		return false, err
	}
	return binding != nil, nil
}

func deleteBinding(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id := d.Id()
	parts := strings.Split(id, "|")
	if len(parts) != 5 {
		return fmt.Errorf("unexpected id '%s' for qpid binding", id)
	}
	b, err := buildBinding(d)
	if err != nil {
		return err
	}
	resp, err := client.DeleteBinding(b)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting binding of destination  '%s' to exchange '%s' using key '%s' on host '%s/%s': %s",
			parts[3], parts[2], parts[4], parts[0], parts[1], resp.Status)
	}
	d.SetId("")
	return nil
}

func updateBinding(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	id := d.Id()
	parts := strings.Split(id, "|")
	if len(parts) != 5 {
		return fmt.Errorf("unexpected id '%s' for qpid binding", id)
	}

	b, err := buildBinding(d)
	if err != nil {
		return err
	}

	resp, err := client.UpdateBinding(b)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid exchange '%s'does not exist on virtual host '%s/%s'", b.Exchange, b.VirtualHostNode, b.VirtualHost)
	}


	return fmt.Errorf("error updating qpid binding of destination '%s' to exchange '%s' using key '%s' on host '%s/%s': %s",
		b.Destination, b.Exchange, b.BindingKey, b.VirtualHostNode, b.VirtualHost, resp.Status)
}
