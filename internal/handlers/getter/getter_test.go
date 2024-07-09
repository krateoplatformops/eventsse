package getter

import (
	"fmt"
	"path"
	"testing"
	"time"
)

func TestKeyWithCompositionId(t *testing.T) {
	basePath := "/krateo.io/events"
	compositionId := "abcdefg"
	creationTimestamp := time.Now()
	eventId := "123456"
	key := creationTimestamp.Format("20060102150405.00")
	key = path.Join(basePath, key, compositionId, eventId)

	fmt.Println(key)
}

func TestKeyWithoutCompositionId(t *testing.T) {
	basePath := "/krateo.io/events"
	creationTimestamp := time.Now()
	eventId := "123456"
	key := creationTimestamp.Format("20060102150405.00")
	key = path.Join(basePath, key, eventId)

	fmt.Println(key)
}
