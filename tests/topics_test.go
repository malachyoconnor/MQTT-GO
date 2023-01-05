package tests

import (
	"MQTT-GO/clients"
	"testing"
)

// Testing if creating a topic map works
func TestInitialization(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x")
	topicStore.PrintTopics()
}

func TestPuttingLowerLevel(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y")

	_, err1 := topicStore.GetMatchingClients("x")
	_, err2 := topicStore.GetMatchingClients("x/y")

	for _, err := range []error{err1, err2} {
		if err != nil {
			t.Error(err)
		}
	}
	topicStore.PrintTopics()
}

// Testing adding an already created top level topic works
func TestDuplicating(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x")
	err := topicStore.AddTopic("x")
	if err != clients.ErrTopicAlreadyExists {
		t.Error("Able to add topic that already exists")
	}
	topicStore.PrintTopics()
}

func TestDuplicatingLowerLevel(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	err := topicStore.AddTopic("x/y/z")

	if err != clients.ErrTopicAlreadyExists {
		t.Error("Able to add topic that already exists")
	}

	topicStore.PrintTopics()
}

func TestAddingMultipleChildren(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	topicStore.AddTopic("x/y/a")
	topicStore.AddTopic("x/y/b")

	_, err1 := topicStore.GetMatchingClients("x/y/z")
	_, err2 := topicStore.GetMatchingClients("x/y/a")
	_, err3 := topicStore.GetMatchingClients("x/y/b")

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			t.Error(err)
		}
	}

	topicStore.PrintTopics()
}

func TestDeletingHigherLevel(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y/z")

	topicStore.DeleteTopic("x")

	if _, err := topicStore.GetMatchingClients("x/y"); err == nil {
		t.Error("Able to access deleted topic")
	}
}

func TestDeletingLowerLevel(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	topicStore.DeleteTopic("x/y")
	topicStore.PrintTopics()

	if _, err := topicStore.GetMatchingClients("x"); err == clients.ErrTopicDoesntExist {
		t.Error("Unable to access base element after child is deleted")
	}

}

func TestDeletingOneChild(t *testing.T) {
	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	topicStore.AddTopic("x/y/a")
	topicStore.AddTopic("x/y/b")
	topicStore.DeleteTopic("x/y/z")

	_, err1 := topicStore.GetMatchingClients("x/y/z")
	_, err2 := topicStore.GetMatchingClients("x/y/a")
	_, err3 := topicStore.GetMatchingClients("x/y/b")

	if err1 != clients.ErrTopicDoesntExist {
		t.Error("Could access deleted element")
	}
	if err2 != nil {
		t.Error(err2)
	}
	if err3 != nil {
		t.Error(err2)
	}

	topicStore.PrintTopics()

}

func TestAddingClientIDs(t *testing.T) {
	t.Deadline()

	topicStore := clients.CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	topicStore.AddTopic("x/y/1")

	topicStore.Put("x/y/z", "abc")
	topicStore.Put("x/y/z", "def")
	topicStore.Put("x/y/1", "def")
	topicStore.PrintTopics()
	res, err := topicStore.GetMatchingClients("x/y/z")

	if res.Head().Value() != "abc" || err != nil {
		t.Error("Value not being added correctly")
	}
}
