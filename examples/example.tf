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
