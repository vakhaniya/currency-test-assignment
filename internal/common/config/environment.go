package config

type Environment string

const (
	Local      Environment = "Local"
	Production Environment = "Production"
)

func (e Environment) IsValid() bool {
	return e == Local || e == Production
}
