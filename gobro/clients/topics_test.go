package clients

import (
	"fmt"
	"testing"

	"golang.org/x/exp/slices"
)

func testErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

// Testing if creating a topic map works
func TestInitialization(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x"))
	topicStore.PrintTopics()
}

func TestPuttingLowerLevel(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.Put("x", "test1"))
	testErr(t, topicStore.Put("x/y", "test2"))

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
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x"))
	err := topicStore.AddTopic("x")
	if err != ErrTopicAlreadyExists {
		t.Error("Able to add topic that already exists")
	}
	topicStore.PrintTopics()
}

func TestDuplicatingLowerLevel(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x/y/z"))
	err := topicStore.AddTopic("x/y/z")

	if err != ErrTopicAlreadyExists {
		t.Error("Able to add topic that already exists")
	}

	topicStore.PrintTopics()
}

func TestAddingMultipleChildren(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.Put("x/y/z", "test1"))
	testErr(t, topicStore.Put("x/y/a", "test2"))
	testErr(t, topicStore.Put("x/y/b", "test3"))

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
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x/y/z"))

	testErr(t, topicStore.Delete("x"))

	if _, err := topicStore.GetMatchingClients("x/y"); err == nil {
		t.Error("Able to access deleted topic")
	}
}

func TestDeletingLowerLevel(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x/y/z"))
	testErr(t, topicStore.Delete("x/y"))
	topicStore.PrintTopics()

	if _, err := topicStore.GetMatchingClients("x"); err == ErrTopicDoesntExist {
		t.Error("Unable to access base element after child is deleted")
	}
}

func TestDeletingOneChild(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x/y/z"))

	testErr(t, topicStore.Put("x/y/z", "abc"))
	testErr(t, topicStore.Put("x/y/z", "def"))
	testErr(t, topicStore.Put("x/y/a", "test1"))
	testErr(t, topicStore.Put("x/y/b", "test2"))

	testErr(t, topicStore.Delete("x/y/z"))

	a, err1 := topicStore.GetMatchingClients("x/y/z")
	b, err2 := topicStore.GetMatchingClients("x/y/a")
	c, err3 := topicStore.GetMatchingClients("x/y/b")

	if err1 != ErrTopicDoesntExist || a != nil {
		t.Error("Could access deleted element")
	}
	if err2 != nil || !b.Contains("test1") {
		t.Error(err2)
	}
	if err3 != nil || !c.Contains("test2") {
		t.Error(err2)
	}

}

func TestAddingClientIDs(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.AddTopic("x/y/z"))
	testErr(t, topicStore.AddTopic("x/y/1"))

	testErr(t, topicStore.Put("x/y/z", "abc"))
	testErr(t, topicStore.Put("x/y/z", "def"))
	testErr(t, topicStore.Put("x/y/1", "def"))
	topicStore.PrintTopics()
	res, err := topicStore.GetMatchingClients("x/y/z")

	if res.Head().Value() != "abc" || err != nil {
		t.Error("Value not being added correctly")
	}
}

func TestDuplicatesAreRemoved(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.Put("x/y/1", "abc"))
	testErr(t, topicStore.Put("x/y/2", "abc"))
	testErr(t, topicStore.Put("x/y/3", "abc"))

	result, _ := topicStore.GetMatchingClients("x/y/#")
	if result.Size() != 1 || result.Head().Value() != "abc" {
		t.Error("Duplicates are not being removed correctly")
	}
}

func TestHashWildcard(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.Put("x/y/z", "abc"))
	testErr(t, topicStore.Put("x/#", "xyz"))

	cLL, _ := topicStore.GetMatchingClients("x/y/z")
	clientArr := cLL.GetItems()
	ServerPrintln(clientArr)
	if !slices.Contains(clientArr, "abc") || !slices.Contains(clientArr, "xyz") ||
		len(clientArr) != 2 {
		t.Error("Didn't find correct clients")
	}
}

func TestPlusWildcard(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.Put("x/y/z", "1"))
	testErr(t, topicStore.Put("x/+/m", "2"))
	testErr(t, topicStore.Put("x/y/c", "3"))

	topicStore.PrintTopics()

	cLL, err := topicStore.GetMatchingClients("x/a/m")

	if err != nil {
		fmt.Println(err)
	}

	if cLL.Size() != 1 || cLL.Head().Value() != "2" {
		t.Error("+ didn't work correctly.")
	}
}

func TestPlusWildcardFurther(t *testing.T) {
	topicStore := CreateTopicTrie()
	testErr(t, topicStore.Put("x/y/z", "1"))
	testErr(t, topicStore.Put("x/M/+", "2"))
	testErr(t, topicStore.Put("x/y/z", "3"))

	cLL, err := topicStore.GetMatchingClients("x/M/z")

	testErr(t, err)
	fmt.Println(cLL.GetItems())

	if cLL.Size() != 1 {
		t.Error("+ didn't work correctly.")
	}
}
