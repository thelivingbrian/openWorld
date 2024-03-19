package main

import (
	"regexp"
	"time"
)

func signUpPage() string {
	return `<h2>Bloop World is currently under development.</h2>`
	/*
		OLD:

		// Trigger back link with backspace
		return `
		<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing"">
			<div>
				<label>Email:</label>
				<input class="retro-input" type="text" name="email" value=""><br />
				<label>Username:</label>
				<input class="retro-input" type="text" name="username" value=""><br />
				<label>Password:</label>
				<input class="retro-input" type="text" name="password" value=""><br />
				<a id="link_submit" href="#">Submit</a><br />
				<a id="link_back" href="/">Back</a>
			</div>
		</form>
		`
	*/
}

func signInPage() string {
	return `
	<form hx-post="/signin" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func (db *DB) newUser(email string, username string, hashword string) string {
	if !isEmailValid(email) {
		return invalidEmailHTML() // Use template to avoid duplication
	}
	user := User{Email: email, Verified: true, Username: username, Hashword: hashword, Created: time.Now()}
	err := db.newAccount(user)
	if err != nil {
		return failedToCreateHTML()
	}
	return "<h1>Success</h1>"
}

func failedToCreateHTML() string {
	return `
	<h2> Username or Email unavailable  </h2>
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func invalidEmailHTML() string {
	return `
	<h3 style='color:red'> Invalid Email. </h3>
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func passwordTooShortHTML() string {
	return `
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<p style='color:red'> Password must have 8 characters. </p>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

func invalidSignin() string {
	return `
	<form hx-post="/signin" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<p style='color:red'> Invalid Sign-in. </p>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}
