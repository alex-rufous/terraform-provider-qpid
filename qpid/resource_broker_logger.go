package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceBrokerLogger() *schema.Resource {

	return &schema.Resource{
		Create: createBrokerLogger,
		Read:   readBrokerLogger,
		Delete: deleteBrokerLogger,
		Update: updateBrokerLogger,
		Exists: existsBrokerLogger,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of broker logger",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of broker logger",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "File" || value == "Console" || value == "BrokerLogbackSocket" || value == "Memory" || value == "Syslog" || value == "JDBC"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid broker logger type value : '%v'", v))
					}

					return
				},
			},

			"description": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},

			"durable": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					_, keySet := d.GetOk(k)
					return !keySet && (old == "true" || new == "true")
				},
			},

			"context": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},

			"virtual_host_log_event_excluded": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
			},

			// File
			"file_name": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"roll_daily": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"roll_on_restart": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"compress_old_files": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"max_history": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"max_file_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"layout": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"console_stream_target", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},

			// Console
			"console_stream_target": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},

			// BrokerLogbackSocket
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"remote_host": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"reconnection_delay": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"include_caller_data": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"mapped_diagnostic_context": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},
			"context_properties": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "max_records"},
			},

			//Syslog
			"syslog_host": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "remote_host",
					"reconnection_delay", "include_caller_data", "include_caller_data", "mapped_diagnostic_context",
					"context_properties", "connection_url", "connection_pool_type", "username", "password",
					"table_name_prefix", "max_records"},
			},
			"suffix_pattern": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "remote_host",
					"reconnection_delay", "include_caller_data", "include_caller_data", "mapped_diagnostic_context",
					"context_properties", "connection_url", "connection_pool_type", "username", "password",
					"table_name_prefix", "max_records"},
			},
			"stack_trace_pattern": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "remote_host",
					"reconnection_delay", "include_caller_data", "include_caller_data", "mapped_diagnostic_context",
					"context_properties", "connection_url", "connection_pool_type", "username", "password",
					"table_name_prefix", "max_records"},
			},
			"throwable_excluded": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "layout", "console_stream_target", "remote_host",
					"reconnection_delay", "include_caller_data", "include_caller_data", "mapped_diagnostic_context",
					"context_properties", "connection_url", "connection_pool_type", "username", "password",
					"table_name_prefix", "max_records"},
			},

			// JDBC
			"connection_url": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "layout",
					"console_stream_target"},
			},
			"connection_pool_type": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "layout",
					"console_stream_target"},
			},
			"username": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "layout",
					"console_stream_target"},
			},
			"password": {
				Type:      schema.TypeString,
				Default:   nil,
				Optional:  true,
				Sensitive: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "layout",
					"console_stream_target"},
			},
			"table_name_prefix": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "layout",
					"console_stream_target"},
			},

			// Memory
			"max_records": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"max_history", "compress_old_files", "roll_on_restart", "roll_daily",
					"file_name", "max_file_size", "port", "remote_host", "reconnection_delay",
					"include_caller_data", "mapped_diagnostic_context", "context_properties", "syslog_host",
					"suffix_pattern", "stack_trace_pattern", "throwable_excluded", "connection_url",
					"connection_pool_type", "username", "password", "table_name_prefix", "layout",
					"console_stream_target"},
			},
		},
	}
}

func createBrokerLogger(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toBrokerLoggerAttributes(d)
	resp, err := client.CreateBrokerLogger(attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetBrokerLogger(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid broker logger'%s': %s", name, resp.Status)
}

func toBrokerLoggerAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceBrokerLogger().Schema, "rule")

	value, exists := d.GetOk("rule")
	if exists {
		var val, expected = value.([]interface{})
		if expected && val != nil {
			var items = make([]map[string]interface{}, len(val))
			for i, v := range val {
				p := v.(map[string]interface{})
				items[i] = *createMapWithKeysInCameCase(&p)
			}
			(*attributes)["rules"] = items
		}
	} else {
		oldValue, newValue := d.GetChange("rule")
		if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", newValue) {
			(*attributes)["rules"] = nil
		}
	}
	return attributes
}

func readBrokerLogger(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetBrokerLogger(name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		return nil
	}

	err = applyResourceAttributes(d, attributes, "rule")
	if err != nil {
		return err
	}

	_, keySet := d.GetOk("rule")
	value, attributeSet := (*attributes)[("rules")]
	if keySet || attributeSet {
		if value != nil {
			rules, validType := value.([]interface{})

			if validType {
				val := make([]map[string]interface{}, len(rules))
				for idx, v := range rules {
					p := v.(map[string]interface{})
					val[idx] = (*createMapWithKeysUnderscored(&p))
				}
				value = val
			} else {
				value = nil
			}
		}

		err = d.Set("rule", value)
		if err != nil {
			return err
		}
	}
	return nil
}

func existsBrokerLogger(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetBrokerLogger(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deleteBrokerLogger(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeleteBrokerLogger(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid broker logger '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateBrokerLogger(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toBrokerLoggerAttributes(d)

	resp, err := client.UpdateBrokerLogger(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  broker logger '%s' does not exist", name)
	}

	return fmt.Errorf("error updating qpid broker logger '%s': %s", name, resp.Status)
}
