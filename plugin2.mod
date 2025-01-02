package main

type Plugin2 struct{}

func (p Plugin2) Execute() string {
    return "Plugin 2 ejecutado"
}

func init() {
    RegisterPlugin("plugin2", Plugin2{})
}