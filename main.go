package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	client_id := flag.String("token", "", "Reddit Client Id")
	client_secret := flag.String("secret", "", "Reddit Secret")
	subs_path := flag.String("subs", "finance_subs.list", "Path of file containing Subreddits and Corresponding flairs")
	flag.Parse()
	client, err := GetClient(
		*client_id,
		*client_secret,
	)
	if err != nil {
		panic(err)
	}
	rawPathBytes, err := os.ReadFile(*subs_path)
	if err != nil {
		panic(err)
	}
	for _, entry := range strings.Split(string(rawPathBytes), "\n") {
		row := strings.Split(entry, " ")
		fmt.Println(row)
		if len(row) == 1 {
			posts, err := client.GetPostsByFlair(row[0], []string{})
			if err != nil {
				panic(err)
			}
			for _, post := range posts {
				fmt.Printf("%+#v\n", post)
			}
		} else if len(row) == 0 {
			continue
		} else {
			flairs := strings.Split(row[1], ",")
			posts, err := client.GetPostsByFlair(row[0], flairs)
			if err != nil {
				panic(err)
			}
			for _, post := range posts {
				fmt.Printf("%+#v\n", post)
			}
		}
	}
}
