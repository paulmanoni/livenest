package main

import (
	"fmt"

	"github.com/paulmanoni/livenest/liveview"
)

// UserForm with validation tags
type UserForm struct {
	Username string `form:"label:Username;placeholder:Enter username" validate:"required;min:3;max:20"`
	Email    string `form:"label:Email Address;type:email;placeholder:your@email.com" validate:"required;email"`
	Password string `form:"label:Password;type:password" validate:"required;min:8"`
	Age      string `form:"label:Age;type:number;placeholder:18" validate:"required;min:13;max:120"`
	Bio      string `form:"label:Bio;type:textarea;rows:4;placeholder:Tell us about yourself" validate:"max:500"`
	Terms    bool   `form:"label:I accept the terms and conditions" validate:"required"`
}

func NewUserForm() *liveview.FormComponent[UserForm] {
	return liveview.NewFormComponent[UserForm]("👤 User Registration").
		OnSubmit(func(socket *liveview.Socket, data *UserForm) error {
			fmt.Printf("User registered: %+v\n", data)
			socket.PutFlash("success", fmt.Sprintf("Welcome, %s!", data.Username))
			return nil
		})
}

// ContactForm with validation
type ContactForm struct {
	Name    string `form:"label:Your Name;placeholder:John Doe" validate:"required;min:2"`
	Email   string `form:"label:Email;type:email" validate:"required;email"`
	Subject string `form:"label:Subject;placeholder:What is this about?" validate:"required;min:5"`
	Message string `form:"label:Message;type:textarea;rows:6" validate:"required;min:10;max:1000"`
}

func NewContactForm() *liveview.FormComponent[ContactForm] {
	return liveview.NewFormComponent[ContactForm]("✉️ Contact Us").
		OnSubmit(func(socket *liveview.Socket, data *ContactForm) error {
			fmt.Printf("Contact: %s <%s> - %s\n", data.Name, data.Email, data.Subject)
			return nil
		})
}

// ProductReview with validation
type ProductReview struct {
	Rating    string `form:"label:Rating (1-5);type:number;placeholder:5" validate:"required;min:1;max:5"`
	Title     string `form:"label:Review Title;placeholder:Summarize your experience" validate:"required;min:3;max:100"`
	Review    string `form:"label:Your Review;type:textarea;rows:6" validate:"required;min:20;max:2000"`
	Recommend bool   `form:"label:I would recommend this product"`
}

func NewProductReview() *liveview.FormComponent[ProductReview] {
	return liveview.NewFormComponent[ProductReview]("⭐ Product Review").
		OnSubmit(func(socket *liveview.Socket, data *ProductReview) error {
			fmt.Printf("Review: %s - Rating: %s\n", data.Title, data.Rating)
			return nil
		})
}

// LoginForm - minimal example
type LoginForm struct {
	Email    string `form:"label:Email;type:email" validate:"required;email"`
	Password string `form:"label:Password;type:password" validate:"required;min:6"`
}

func NewLoginForm() *liveview.FormComponent[LoginForm] {
	return liveview.NewFormComponent[LoginForm]("🔐 Login").
		OnSubmit(func(socket *liveview.Socket, data *LoginForm) error {
			fmt.Printf("Login: %s\n", data.Email)
			// Your authentication logic here
			return nil
		})
}
