package main

import (
	"fmt"
	"time"

	user "cilium.com/examples/internal/pkg1"
)

func main() {
	fmt.Println("Hello, World!")

	user1 := user.NewUser("Alice1", 31)
	fmt.Println(user1)
	user2 := user.NewUser("Alice2", 32)
	fmt.Println(user2)

	user3 := user.NewUser("Alice3", 33)
	fmt.Println(user3)
	user4 := user.NewUser("Alice4", 34)
	fmt.Println(user4)
	populateUsers(getUsers())
	fmt.Println("hi ", getValue("Bob1"))
	var ip <-chan *user.User

	go func() {
		ip = process(getUsers())
	}()
	time.Sleep(time.Second * 30)

	go read(ip)
	time.Sleep(time.Second * 30)
	select {}

}

func getUsers() []*user.User {
	return []*user.User{
		user.NewUser("Bob1", 21),
		user.NewUser("Bob2", 22),
		user.NewUser("Bob3", 23),
		user.NewUser("Bob4", 24),
	}
}

func readUsers(users []*user.User) {
	for _, u := range users {
		fmt.Println(u)
	}
}

func populateUsers(users []*user.User) {
	cache = map[string]*user.User{}
	for _, u := range users {
		cache[u.Name] = u
	}
}

var cache map[string]*user.User

func getValue(name string) *user.User {
	return cache[name]
}

func process(users []*user.User) <-chan *user.User {
	results := make(chan *user.User, 5)
	fmt.Println("================read ", results)

	for _, u := range users {
		results <- u
	}
	return results
}

func read(data <-chan *user.User) {
	fmt.Println("================read ", data)
	for user := range data {
		fmt.Println("resd ", *user)
	}
}
