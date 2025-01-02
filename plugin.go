package main

type Plugin interface {
	Execute() string
}

var pluginRepo = make(map[string]Plugin)

func RegisterPlugin(name string, plugin Plugin) {
	pluginRepo[name] = plugin
}

func GetPlugin(name string) (Plugin, bool) {
	plugin, found := pluginRepo[name]
	return plugin, found
}

func ListPlugins() []string {
	keys := make([]string, 0, len(pluginRepo))
	for key := range pluginRepo {
		keys = append(keys, key)
	}
	return keys
}
