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
	time.Sleep(time.Second * 10)

	go read(ip)
	time.Sleep(time.Second * 30)
	fmt.Println("end")
	//select {}

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
	//=====1st issue
	// var e user.User = user.User{}
	// for _, u := range users {
	// 	e = *(u)
	// 	fmt.Println("rest ", *u)
	// 	results <- &e
	// }
	//=====2nd issue
	// var e user.User = user.User{}
	// for _, u := range users {
	// 	e = user.User{}
	// 	e = *(u)
	// 	fmt.Println("rest ", *u)
	// 	results <- &e
	// }
	//=====corrected
	// 	//another way is for _, u := range users {
	//     e := *u // Create a NEW struct variable `e` on every iteration
	//     results <- &e // Send a pointer to this new, unique struct
	// }
	var e *user.User = &user.User{}
	for _, u := range users {
		e = &user.User{}
		*e = *u
		fmt.Println("rest ", *u)
		results <- e
	}
	close(results)
	return results
}

func read(data <-chan *user.User) {
	time.Sleep(10 * time.Second)
	fmt.Println("================read ", data)
	for user := range data {
		fmt.Println("resd ", *user)
	}
}
