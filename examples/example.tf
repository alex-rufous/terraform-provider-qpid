provider "qpid" {
  endpoint = "https://127.0.0.1:8080"
  username = "guest"
  password = "guest"
  model_version = "v7.1"
  skip_cert_verification = true
  certificate_file = "./qpid-cert.pem"
}


# Create a virtual host node 'foo' with initial configuration
resource "qpid_virtual_host_node" "foo" {
  name = "foo"
  type = "JSON"
  virtual_host_initial_configuration = "{\"type\":\"BDB\"}"
}

# Create a virtual host node 'test'
resource "qpid_virtual_host_node" "test" {
  name = "test"
  type = "JSON"
}

# Create BDB_HA virtual host node
resource "qpid_virtual_host_node" "node1" {
  name = "node1"
  type = "BDB_HA"
  address = "localhost:5000"
  helper_address = "localhost:5000"
  group_name = "foo"
  permitted_nodes = ["localhost:5000", "localhost:5001", "localhost:5002"]
}


# Create a virtual host 'test'
resource "qpid_virtual_host" "test" {
  depends_on = [qpid_virtual_host_node.test]
  virtual_host_node = "test"
  name = "test"
  type = "BDB"
  node_auto_creation_policy {
                                   pattern = "auto-creation-queue.*"
                                   created_on_publish = true
                                   created_on_consume = true
                                   node_type = "Queue"
                                   attributes = {"maximum_delivery_attempts"= "2", "alternate_binding" = "{\"destination\": \"bar\"}"}
                            }
  node_auto_creation_policy {
                                   pattern = "auto-creation-topic.*"
                                   created_on_publish = true
                                   created_on_consume = false
                                   node_type = "Exchange"
                                   attributes = {"unroutable_message_behaviour"= "reject"}
                            }
}

# Create a priority queue
resource "qpid_queue" "my-priority-queue" {
  depends_on = [qpid_virtual_host.test]
  virtual_host_node = "test"
  virtual_host = "test"
  name = "my-priority-queue"
  type = "priority"
  priorities = 10
}


# Create a standard queue
resource "qpid_queue" "my-standard-queue" {
  depends_on = [qpid_virtual_host.test]
  virtual_host_node = "test"
  virtual_host = "test"
  name = "my-standard-queue"
  type = "standard"
}

# Create a last value queue
resource "qpid_queue" "my-lvq" {
  depends_on = [qpid_virtual_host.test]
  virtual_host_node = "test"
  virtual_host = "test"
  name = "my-lvq"
  type = "lvq"
  lvq_key = "myKey"
}
# Create a sorted queue
resource "qpid_queue" "my-sorted-queue" {
  depends_on = [qpid_virtual_host.test]
  virtual_host_node = "test"
  virtual_host = "test"
  name = "my-sorted-queue"
  type = "sorted"
  sort_key = "mySortedKey"
}

# Create a  queue 'bar' with default filters
resource "qpid_queue" "bar" {
  depends_on = [qpid_virtual_host.test]
  virtual_host_node = "test"
  virtual_host = "test"
  name = "bar"
  type = "standard"
  maximum_message_ttl = 99999
  minimum_message_ttl     = 88888
  default_filters= "{ \"x-filter-jms-selector\" : { \"x-filter-jms-selector\" : [ \"id>0\" ] } }"
}

# Create a  queue 'foo'
resource "qpid_queue" "foo" {
  depends_on = [qpid_virtual_host.test]
  virtual_host_node = "test"
  virtual_host = "test"
  name = "foo"
  type = "standard"
  maximum_message_ttl = 99999
  minimum_message_ttl     = 88888
  default_filters= "{ \"x-filter-jms-selector\" : { \"x-filter-jms-selector\" : [ \"id>0\" ] } }"
}

# Create a standard queue with alternate binding
resource "qpid_queue" "my-queue" {
  depends_on = [qpid_virtual_host.test, qpid_queue.bar]
  type = "standard"
  virtual_host_node = "test"
  virtual_host = "test"
  name = "my-queue"
  minimum_message_ttl = 10000000
  alternate_binding {
    destination = "bar"
    attributes = {
        "x-filter-jms-selector"= "id>0"
    }
  }
}

resource "qpid_queue" "blah" {
  depends_on = [qpid_virtual_host.test]
  type = "standard"
  virtual_host_node = "test"
  virtual_host = "test"
  name = "blah"
  context = {"foo"= "bar", "bar" = "foo", "one"= "two"}
  maximum_delivery_attempts = 2
}

resource "qpid_exchange" "my_exchange" {
  depends_on = [qpid_virtual_host.test, qpid_queue.blah]
  type = "direct"
  virtual_host_node = "test"
  virtual_host = "test"
  name = "my_exchange"
  alternate_binding {
    destination = "blah"
    attributes = {
        "x-filter-jms-selector"= "id>0"
    }
  }
  unroutable_message_behaviour = "DISCARD"
}

resource "qpid_binding" "bnd" {
  depends_on = [qpid_queue.foo, qpid_exchange.my_exchange]
  virtual_host_node = "test"
  virtual_host = "test"
  binding_key = "bnd"
  destination = "bar"
  exchange = "my_exchange"
  arguments = {
          "x-filter-jms-selector"= "a is not null"
  }
}


resource "qpid_binding" "bnd2" {
  depends_on = [qpid_queue.blah, qpid_exchange.my_exchange]
  virtual_host_node = "test"
  virtual_host = "test"
  binding_key = "bnd2"
  destination = "blah"
  exchange = "my_exchange"
  arguments = {
          "x-filter-jms-selector" = "b=2"
  }
}

resource "qpid_authentication_provider" "auth" {
  type = "Plain"
  name = "auth"
}

resource "qpid_user" "test_user" {
  depends_on = [qpid_authentication_provider.auth]
  name = "test_user"
  type = "managed"
  password = "bar"
  authentication_provider = "auth"
}

resource "qpid_group_provider" "groups" {
  type = "ManagedGroupProvider"
  name = "groups"
}

resource "qpid_group" "admins" {
  type = "ManagedGroup"
  name = "admins"
  group_provider = "groups"
}

resource "qpid_group" "messaging" {
  depends_on = [qpid_group_provider.groups]
  type = "ManagedGroup"
  name = "messaging"
  group_provider = "groups"
}

resource "qpid_group_member" "admin" {
  depends_on = [qpid_group.admins]
  type = "ManagedGroupMember"
  name = "admin"
  group_provider = "groups"
  group = "admins"
}

resource "qpid_group_member" "client" {
  depends_on = [qpid_group.messaging]
  type = "ManagedGroupMember"
  name = "client"
  group_provider = "groups"
  group = "messaging"
}


resource "qpid_access_control_provider" "acl" {
  name = "acl"
  type = "RuleBased"
  priority = 1

  rule {
    outcome = "ALLOW_LOG"
    identity = "guest"
    operation = "ALL"
    object_type = "ALL"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "admins"
    operation = "ALL"
    object_type = "ALL"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "ACCESS"
    object_type = "VIRTUALHOST"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "ACCESS"
    object_type = "MANAGEMENT"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "CONSUME"
    object_type = "ALL"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "PUBLISH"
    object_type = "ALL"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "BIND"
    object_type = "EXCHANGE"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "UNBIND"
    object_type = "EXCHANGE"
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "CREATE"
    object_type = "QUEUE"
    attributes = {
      "TEMPORARY"="true"
      "AUTO_DELETE"="true"
    }
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "DELETE"
    object_type = "QUEUE"
    attributes = {
      "TEMPORARY"="true"
      "AUTO_DELETE"="true"
    }
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "INVOKE"
    object_type = "QUEUE"
    attributes = {
      "METHOD_NAME"="getMessage*"
    }
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "INVOKE"
    object_type = "QUEUE"
    attributes = {
      "METHOD_NAME"="deleteMessages"
    }
  }

  rule {
    outcome = "ALLOW_LOG"
    identity = "messaging"
    operation = "ACCESS_LOGS"
    object_type = "ALL"
  }

  rule {
    outcome = "DENY_LOG"
    identity = "ALL"
    operation = "ALL"
    object_type = "ALL"
  }

}

resource "qpid_key_store" "my_keystore" {
  name = "my_keystore"
  type = "NonJavaKeyStore"
  private_key_url = "data:;base64,MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQC83m//nfqePNdogJSRCXxYRyyaPJkCi/o0IlfGaNtgoFskwUCjJE1Wj7TFKSFVey2vGN3MwgJYSTRHrudiaTZxPYtVyD8YkMd5JipRg7xURvpunKkIvQd9d9Pg1bmzVoKc8M7NmNpZUV219hywLNh02xReZbFmzDy9ec09uP6Czec6GMsKiM9ujUJAMni+EPmBDdgBuFQzzVHAJlQGERptMN47cn57OTes/dyDXvKF/I63Zl20NnRFg5EozgZmhKkyk5BzYHFPfOB+NSS1nhAsZ4h1onjMmDo2N8R2PPSC6KwteaKghd5oT4AJyR3f+eBf3/R3IrnM+wViPn/BQiq3AgMBAAECggEATrWklzKPL0LLwpFTWN5LI78Fp4F5gsYzD2cAjX9FbY9mbHrdJSAL2vcorsHlUmpzL3V9ecegkopvbzBE3Y5bUfYEC0vYf7RWbPaqzC3KXpT16QMArtOYO4Gkmx52tXZoGF+Cz8vTs0VleF+ItSL7Uje61VwsAls7NPt9vStLZdcQs9PiNyF2qRVc2q/csBzJlyJfJWERq9y7hBrwhvQmHtDa0NbjC0CGvRgPk28HDyGwblN/4XUH/xmWrLHSn2sGOD4DsAJJayb256MmvEUL+HUo6rlhbPOG5hgn0vYBJGFSRNEZI/Ocy905bjN/huxsICoUGYUi13q8iuOJr/A3AQKBgQD5smGPu6PJXx55hA1IUsGxGYOxEKn06z9+iotWZZmV2Gq51N9stQBmcj1TPDYZrLMxagDddPKSyQFbm6AIln16+Qd0qLAEqZ0QDq3d+eMS7YoTIbutzeyPLoP74kXhBahzoCB4jPqIG7LruCUEkBA5IkgOxRFsFHeVC5DLCCCLzwKBgQDBoveplVir+p/E1dIvq2iJdHmYsl3KZCngGVm1xepsQhMjCPoQGsWfAvloBtXQnh26IlFr9GI8etgWmT9k+xd5er6v1NTHZ3udTwpdrywjz7nnAbIwdzNbAWmRSNmktqqjRhSlmtkqNnjWMu0BhOPtAkv426xB56B+1UDIyWKkmQKBgQCi8VwnHqy4SSEq7Rh53L9XIa5FivlNwYJywlhBLhX2qf6jfB2847T6JYyNV5p6UK+zDFi6K4nsbc08Cad6UzJZYE8UOsx6jnDXPK0LUPl0rZxP9dBykBHSMemhIry1JisSISlvYZhP37t3hXhqrNRZZFyffsxqukR698wqIgiTEwKBgQC8fJg7uSaxcarn/YM159JASuK6YpWtl0az37lVmawRaVgbeHeCCa1olYqVWmHzSpaBQzqirSaa3LFPfikZcNlu5K6Nlczxtae8ft4GR6fdzCyX0yzSxJV29q7+Pz2seisr9+HNOig+UPva9YODQQplASFWwu6w0HmIPKltSar9sQJ/YFUxftE1HF8SdBFVONfz4WOUYHqr+uv69KKHgWDYHCGP4vgzZRmHs8V7cGUYtoWX0PJGjDa82sTx85iRRJk1tvot+GBd1gxAbot+bKBMjeJ6jJ9IJ1v4Bp+4yGppSD8f8rW7MKkJYrhkWY5ODcuc0hd41C0ARerj+0zAZpq5UA=="
  certificate_url = "data:;base64,MIIC7jCCAdagAwIBAgIRAIOytEG1ugWCCq9v7KzoHV4wDQYJKoZIhvcNAQELBQAwEjEQMA4GA1UEChMHRm9vIE9yZzAgFw0yMDAyMjQxNDAwNThaGA8yMTIwMDEzMTE0MDA1OFowEjEQMA4GA1UEChMHRm9vIE9yZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALzeb/+d+p4812iAlJEJfFhHLJo8mQKL+jQiV8Zo22CgWyTBQKMkTVaPtMUpIVV7La8Y3czCAlhJNEeu52JpNnE9i1XIPxiQx3kmKlGDvFRG+m6cqQi9B3130+DVubNWgpzwzs2Y2llRXbX2HLAs2HTbFF5lsWbMPL15zT24/oLN5zoYywqIz26NQkAyeL4Q+YEN2AG4VDPNUcAmVAYRGm0w3jtyfns5N6z93INe8oX8jrdmXbQ2dEWDkSjOBmaEqTKTkHNgcU984H41JLWeECxniHWieMyYOjY3xHY89ILorC15oqCF3mhPgAnJHd/54F/f9Hciucz7BWI+f8FCKrcCAwEAAaM9MDswDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMBQGA1UdEQQNMAuCCWxvY2FsaG9zdDANBgkqhkiG9w0BAQsFAAOCAQEAHAEbEKY2pco3brCwc7slCMgwDjuYwRnGU7uwfIvoe8KHRZDzmBm1/JjqCWVFUhE7sJDg4/aeomd3TEObVH+X2+QScws4aniRGMC6vNAqKn1JX1M54sixFHO7LIArVNmEc/5H8mMfKJ2fo4Ih9PtRA69530713gmugdYla8D5VrIQJKPUhEJyLuYdTqmMfJEDJLtlRgjxzlgb6DENhsGZIM4/kJVJ5Tvp1kvqLo9Z6ctnMoZzNwPSj/b/KH4KGf5jm5AcVynpAztKPbCCFX6s6pMDb0SxEgeBIJSLx7YIcykb4NW85GPOVipW1NoeAlcD+rKqtGroxCPdFMDQCn8c6g=="
}
