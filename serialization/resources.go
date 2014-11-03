package serialization

import (
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func LogConfigToModel(config receptor.LogConfig) models.LogConfig {
	return models.LogConfig{
		Guid:       config.Guid,
		SourceName: config.SourceName,
	}
}

func LogConfigFromModel(config models.LogConfig) receptor.LogConfig {
	return receptor.LogConfig{
		Guid:       config.Guid,
		SourceName: config.SourceName,
	}
}

func EnvironmentVariablesToModel(envVars []receptor.EnvironmentVariable) []models.EnvironmentVariable {
	out := make([]models.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func EnvironmentVariablesFromModel(envVars []models.EnvironmentVariable) []receptor.EnvironmentVariable {
	out := make([]receptor.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func PortMappingToModel(ports []receptor.PortMapping) []models.PortMapping {
	if len(ports) == 0 {
		return nil
	}
	out := make([]models.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = val.ContainerPort
		out[i].HostPort = val.HostPort
	}
	return out
}

func PortMappingFromModel(ports []models.PortMapping) []receptor.PortMapping {
	if len(ports) == 0 {
		return nil
	}
	out := make([]receptor.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = val.ContainerPort
		out[i].HostPort = val.HostPort
	}
	return out
}
