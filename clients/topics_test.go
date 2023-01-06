package clients

import (
	"fmt"
	"testing"

	"golang.org/x/exp/slices"
)

// Testing if creating a topic map works
func TestInitialization(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.AddTopic("x")
	topicStore.PrintTopics()
}

func TestPuttingLowerLevel(t *testing.T) {
	topicStore := CreateTopicMap()
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
	topicStore := CreateTopicMap()
	topicStore.AddTopic("x")
	err := topicStore.AddTopic("x")
	if err != ErrTopicAlreadyExists {
		t.Error("Able to add topic that already exists")
	}
	topicStore.PrintTopics()
}

func TestDuplicatingLowerLevel(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	err := topicStore.AddTopic("x/y/z")

	if err != ErrTopicAlreadyExists {
		t.Error("Able to add topic that already exists")
	}

	topicStore.PrintTopics()
}

func TestAddingMultipleChildren(t *testing.T) {
	topicStore := CreateTopicMap()
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
	topicStore := CreateTopicMap()
	topicStore.AddTopic("x/y/z")

	topicStore.Delete("x")

	if _, err := topicStore.GetMatchingClients("x/y"); err == nil {
		t.Error("Able to access deleted topic")
	}
}

func TestDeletingLowerLevel(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	topicStore.Delete("x/y")
	topicStore.PrintTopics()

	if _, err := topicStore.GetMatchingClients("x"); err == ErrTopicDoesntExist {
		t.Error("Unable to access base element after child is deleted")
	}

}

func TestDeletingOneChild(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.AddTopic("x/y/z")
	topicStore.AddTopic("x/y/a")
	topicStore.AddTopic("x/y/b")
	topicStore.Delete("x/y/z")

	_, err1 := topicStore.GetMatchingClients("x/y/z")
	_, err2 := topicStore.GetMatchingClients("x/y/a")
	_, err3 := topicStore.GetMatchingClients("x/y/b")

	if err1 != ErrTopicDoesntExist {
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
	topicStore := CreateTopicMap()
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

func TestDuplicatesAreRemoved(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.Put("x/y/1", "abc")
	topicStore.Put("x/y/2", "abc")
	topicStore.Put("x/y/3", "abc")

	result, _ := topicStore.GetMatchingClients("x/y/#")
	if result.Size != 1 || result.Head().Value() != "abc" {
		t.Error("Duplicates are not being removed correctly")
	}

}

func TestHashWildcard(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.Put("x/y/z", "abc")
	topicStore.Put("x/#", "xyz")

	cLL, _ := topicStore.GetMatchingClients("x/y/z")
	clientArr := cLL.GetItems()
	fmt.Println(clientArr)
	if !slices.Contains(clientArr, "abc") || !slices.Contains(clientArr, "xyz") ||
		len(clientArr) != 2 {
		t.Error("Didn't find correct clients")
	}

}

func TestPlusWildcard(t *testing.T) {
	topicStore := CreateTopicMap()
	topicStore.Put("x/y/z", "abc")
	topicStore.Put("x/+/m", "xyz")
	topicStore.Put("x/y/c", "xyz")

	cLL, _ := topicStore.GetMatchingClients("x/a/m")

	if cLL.Size != 1 || cLL.Head().Value() != "xyz" {
		t.Error("+ didn't work correctly.")
	}

}
