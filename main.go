package main

import (
	"go-practices/tree"
	"fmt"
	"sync"
	"go-practices/crawler"
)

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	WalkTree(t, ch)
	close(ch)
}

func WalkTree(t *tree.Tree, ch chan int) {

	if t.Left != nil {
		WalkTree(t.Left, ch)
	}

	ch <- t.Value

	if t.Right != nil {
		WalkTree(t.Right, ch)
	}
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {

	ch1, ch2 := make(chan int), make(chan int)

	go Walk(t1, ch1)
	go Walk(t2, ch2)

	for i := range ch1 {
		if i != <-ch2 {
			return false
		}
	}

	return true
}



type urlCache struct {
	sync.Mutex
	mapping map[string]bool
}

var cache = urlCache{mapping: make(map[string]bool)}

func CrawlParallel(url string, depth int, fetcher crawler.Fetcher, terminal chan string) {
	if depth <= 0 {
		terminal <- "destination"
		return
	}

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		terminal <- "termination"
		return
	}

	fmt.Printf("found: %s %q\n", url, body)

	for _, u := range urls {
		if _, prs := cache.mapping[u]; !prs {

			cache.Lock()

			cache.mapping[u] = true

			cache.Unlock()

			go CrawlParallel(u, depth-1, fetcher, terminal)
		}
	}

	terminal <- "end"
	return
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher crawler.Fetcher) {

	terminal := make(chan string, depth) //set the number of the rounds by depth
	go CrawlParallel(url, depth, fetcher, terminal)

	for i := depth; i > 0; i-- {
		<-terminal
	}

	for u := range cache.mapping {
		fmt.Println("Url fetched:" + u)
	}

	return
}

func main() {

	ch := make(chan int)
	tree1 := tree.New(1)
	go Walk(tree1, ch)

	for i := range ch {
		fmt.Println(i)
	}

	fmt.Println("Should return true:", Same(tree.New(1), tree.New(1)))
	fmt.Println("Should return false:", Same(tree.New(1), tree.New(2)))

	Crawl("https://golang.org/", 4, fetcher)
}


// fetcher is a populated fakeFetcher.
var fetcher = crawler.FakeFetcher{
	"https://golang.org/": &crawler.FakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &crawler.FakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &crawler.FakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &crawler.FakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}