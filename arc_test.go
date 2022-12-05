package gcache

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"testing"
	"time"
)

func TestARCGet(t *testing.T) {
	size := 1000
	gc := buildTestCache(t, TYPE_ARC, size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingARCGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildTestLoadingCache(t, TYPE_ARC, size, loader), numbers)
}

func TestARCLength(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_ARC, 2, time.Millisecond)
	gc.Get("test1")
	gc.Get("test2")
	gc.Get("test3")
	length := gc.Len(true)
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
	time.Sleep(time.Millisecond)
	gc.Get("test4")
	length = gc.Len(true)
	expectedLength = 1
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
}

func TestARCEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := cacheSize + 1
	gc := buildTestLoadingCache(t, TYPE_ARC, cacheSize, loader)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestARCPurgeCache(t *testing.T) {
	cacheSize := 10
	purgeCount := 0
	gc := New(cacheSize).
		ARC().
		LoaderFunc(loader).
		PurgeVisitorFunc(func(k, v interface{}) {
			purgeCount++
		}).
		Build()

	for i := 0; i < cacheSize; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	gc.Purge()

	if purgeCount != cacheSize {
		t.Errorf("failed to purge everything")
	}
}

func TestARCGetIFPresent(t *testing.T) {
	testGetIFPresent(t, TYPE_ARC)
}

func TestARCHas(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_ARC, 2, 10*time.Millisecond)

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			gc.Get("test1")
			gc.Get("test2")

			if gc.Has("test0") {
				t.Fatal("should not have test0")
			}
			if !gc.Has("test1") {
				t.Fatal("should have test1")
			}
			if !gc.Has("test2") {
				t.Fatal("should have test2")
			}

			time.Sleep(20 * time.Millisecond)

			if gc.Has("test0") {
				t.Fatal("should not have test0")
			}
			if gc.Has("test1") {
				t.Fatal("should not have test1")
			}
			if gc.Has("test2") {
				t.Fatal("should not have test2")
			}
		})
	}
}

func TestARCRandom(t *testing.T) {
	const size = 100
	gc := buildTestCache(t, TYPE_ARC, size)

	rando := rand.New(rand.NewSource(5))

	hash := fnv.New64a()

	for _, nKeys := range []int{size / 2, size, size * 2} {
		var keys []int
		for i := 0; i < nKeys; i++ {
			keys = append(keys, rand.Int())
		}

		for i := 0; i < 100_000; i++ {
			if rando.Intn(2) == 0 {
				k := keys[rando.Intn(len(keys))]
				gc.Set(k, k+1)
			}
			if rando.Intn(2) == 0 {
				k := keys[rando.Intn(len(keys))]
				v, err := gc.Get(k)
				hash.Write([]byte(fmt.Sprint(k, v, err)))
			}
		}
	}

	got := hash.Sum64()
	if got != 16111254380956458052 {
		t.Fatal(got)
	}
}
