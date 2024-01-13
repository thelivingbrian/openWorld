package main

import (
	"regexp"
	"time"
)

func signUpPage() string {
	return `
	<form hx-post="/signup" hx-target="#landing"">
		<div>
			<label>Email:</label>
			<input type="text" name="email" value=""><br />
			<label>Username:</label>
			<input type="text" name="username" value=""><br />
			<label>Password:</label>
			<input type="text" name="password" value=""><br />
		</div>
		<button>Sign Up!</button>
	</form>
	`
}

func signInPage() string {
	return `
	<form hx-post="/signin" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input type="text" name="email" value=""><br />
			<label>Password:</label>
			<input type="text" name="password" value=""><br />
		</div>
		<button>Sign In</button>
	</form>
	`
}

func (app *App) newUser(email string, username string, hashword string) string {
	if !isEmailValid(email) {
		return invalidEmailHTML()
	}
	user := User{Email: email, Verified: true, Username: username, Hashword: hashword, Created: time.Now()}
	err := newAccount(app.db, user)
	if err != nil {
		return failedToCreateHTML()
		//log.Fatal(err)
	}
	return "<h1>Success</h1>"
}

func failedToCreateHTML() string {
	return `
	<h2> Username or Email unavailable  </h2>
	<form hx-post="/signup" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input type="text" name="email" value=""><br />
			<label>Username:</label>
			<input type="text" name="username" value=""><br />
			<label>Password:</label>
			<input type="text" name="password" value=""><br />
		</div>
		<button>Sign Up!</button>
	</form>
	`
}

func invalidEmailHTML() string {
	return `
	<h2 style='color:red'> Invalid Email. </h2>
	<form hx-post="/signup" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input type="text" name="email" value=""><br />
			<label>Username:</label>
			<input type="text" name="username" value=""><br />
			<label>Password:</label>
			<input type="text" name="password" value=""><br />
		</div>
		<button>Sign Up!</button>
	</form>
	`
}

func passwordTooShortHTML() string {
	return `
	<h2 style='color:red'> Password must have 8 characters. </h2>
	<form hx-post="/signup" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input type="text" name="email" value=""><br />
			<label>Username:</label>
			<input type="text" name="username" value=""><br />
			<label>Password:</label>
			<input type="text" name="password" value=""><br />
		</div>
		<button>Sign Up!</button>
	</form>
	`
}

func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

func invalidSignin() string {
	return `
	<h2 style='color:red'> Invalid Signin. </h2>
	<form hx-post="/signin" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input type="text" name="email" value=""><br />
			<label>Password:</label>
			<input type="text" name="password" value=""><br />
		</div>
		<button>Sign In</button>
	</form>
	`
}
