Chapter 1 and 2: Common Mistakes in Go Concurrency

1. Variable shadowing
2. No getters and setters
3. keep interfaces minimal (The bigger the interface, the weaker the abstraction.)
4. Abstractions should be discovered not created beforehand (Exceptions exists always)
5. interfaces should be preferred to be at consumer side
6. accept interfaces and return structs if possible
7. empty struct size is zero, any {} got two ptrs
7. channels, maps, functions and slices are reference types
8. maps, functions and slices are non comparable
9. structs , arrays containg non comparable field are non comparable
10. Declaring map  Map[any]int is bad, Ex: key should be comparable type, but any{} implies we can use []byte as key which is not comparable
11. Similar to interfaces use of generics should be discovered not created beforehand
12. you can't add new type parameters to the method, you must define a top-level generic function instead 
    ```
    type MyData[T any] struct {
        value T
    }
    // Incorrect: trying to add new type parameter U to method Map
    func (d MyData[T]) Map[U any](f func(T) U) MyData[U] {
        return MyData[U]{value: f(d.value)}
    }
    // Correct: define a top-level generic function instead
    func MapMyData[T any, U any](d MyData[T], f func(T) U) MyData[U] {
        return MyData[U]{value: f(d.value)}
    }
            
    ```
13. Check Options and builder patterns
14. In Go, there is no concept of subpackages. However, we can decide to organize packages within subdirectories.
15. Instead of a utility package, we should create an expressive package name such as stringset
16. Avoid circular dependencies between packages
17. 
