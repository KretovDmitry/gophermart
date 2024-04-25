package request

// Register defines parameters for Register.
type Register struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Login defines parameters for Login.
type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
