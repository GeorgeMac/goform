goform
======

## Description

Unmarshal and validate form values from a http request in to a target struct

## Example Usage

```go
package main

import (
    "net/http"
    "github.com/GeorgeMac/goform"
)
type UserForm struct {
    Name  string    `form:"name"`
    Email string    `form:"email"`
    DOB   time.Time `form:"dob"`
}

type EmailField string

func (_ EmailField) Validate(v interface{}) error {
    if !strings.Contains(v.(string), "@") {
        return errors.New("Email Must Contain @ Symbol")
    }

    return nil
}

func main() {
    http.Handle("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        user := &User{}

        if err := goform.Unmarshal(r, user); err != nil {
            if err, ok := err.(goform.ValidationError); ok {
                if err := json.NewEncoder(w).Encode(err); err != nil {
                    http.Error(w, "Something Went Wrong", http.StatusInternalServerError)
                    return
                }
                w.WriteHeader(http.StatusBadRequest)
                return
            }
            http.Error(w, "Something Else Went Wrong", http.StatusInternalServerError)
        }

        // persist user or whatever...
    }))
    http.ListenAndServe(":8080", nil)
}
```
