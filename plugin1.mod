package main

type Plugin1 struct{}

func (p Plugin1) Execute() string {
    return "Plugin 1 ejecutado"
}

func init() {
    RegisterPlugin("plugin1", Plugin1{})
}