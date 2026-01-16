package auth_case

type SessionTracker struct {
	JTI      string
	UserID   string
	Token    string
	Device   string
	UserAgen string
	IP       string
	LoginAt  string
}
