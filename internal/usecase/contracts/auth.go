package contracts

type Authorizer interface {
	Generate(string) (string, error)
	Check(string) bool
}
