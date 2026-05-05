package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// LoadFile loads configuration from YAML file, .env file, and environment variables.
// Priority: env vars > .env file > YAML file > defaults
func (c *Config) LoadFile(path string) error {
	// 1. Load .env file if exists (same directory as config or CWD)
	envPath := filepath.Join(filepath.Dir(path), ".env")
	if _, err := os.Stat(envPath); err == nil {
		_ = godotenv.Load(envPath)
	} else {
		_ = godotenv.Load(".env")
	}

	// 2. Load YAML config file
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// No config file is okay - rely on env vars and defaults
			c.applyEnvOverrides()
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	// 3. Apply YAML values to Config
	c.applyRaw(raw)

	// 4. Apply environment variable overrides (highest priority)
	c.applyEnvOverrides()

	return nil
}

// Save writes the current configuration back to the config file.
func (c *Config) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := yaml.Marshal(c.toMap())
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	header := []byte("# ClawSec Configuration File\n# Generated automatically\n\n")
	if err := os.WriteFile(path, append(header, data...), 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// applyRaw maps a raw map to Config fields using reflection and yaml tags.
func (c *Config) applyRaw(raw map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fv := val.Field(i)

		// Skip unexported and special fields
		if !fv.CanSet() || field.Name == "mu" || field.Name == "raw" {
			continue
		}

		tag := field.Tag.Get("yaml")
		if tag == "" || tag == "-" {
			continue
		}
		key := strings.Split(tag, ",")[0]

		v, ok := raw[key]
		if !ok {
			continue
		}

		setField(fv, v)
	}

	// Handle nested "ai" section
	if aiRaw, ok := raw["ai"].(map[string]interface{}); ok {
		if v, ok := aiRaw["enabled"].(bool); ok {
			c.AIEnabled = v
		}
		if v, ok := aiRaw["endpoint"].(string); ok {
			c.AIEndpoint = v
		}
		if v, ok := aiRaw["model"].(string); ok {
			c.AIModel = v
		}
		if v, ok := aiRaw["api_key"].(string); ok {
			c.AIAPIKey = v
		}
	}

	// Handle nested "products" section
	if prodRaw, ok := raw["products"].(map[string]interface{}); ok {
		for name, pcfg := range prodRaw {
			if pm, ok := pcfg.(map[string]interface{}); ok {
				pc := ProductConfig{}
				if v, ok := pm["url"].(string); ok {
					pc.URL = v
				}
				if v, ok := pm["api_key"].(string); ok {
					pc.APIKey = v
				}
				if v, ok := pm["token"].(string); ok {
					pc.Token = v
				}
				if v, ok := pm["username"].(string); ok {
					pc.Username = v
				}
				if v, ok := pm["password"].(string); ok {
					pc.Password = v
				}
				if v, ok := pm["insecure"].(bool); ok {
					pc.Insecure = v
				}
				c.Products[name] = pc
			}
		}
	}
}

// applyEnvOverrides applies CLAWSEC_* environment variables using reflection.
// Maps: CLAWSEC_OUTPUT_FORMAT -> OutputFormat, CLAWSEC_AI_ENDPOINT -> AIEndpoint, etc.
func (c *Config) applyEnvOverrides() {
	c.mu.Lock()
	defer c.mu.Unlock()

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	prefix := "CLAWSEC_"

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fv := val.Field(i)

		if !fv.CanSet() || field.Name == "mu" || field.Name == "raw" {
			continue
		}

		envKey := prefix + toEnvKey(field.Name)
		if envVal := os.Getenv(envKey); envVal != "" {
			setFieldFromString(fv, envVal)
		}
	}

	// Nested AI settings
	if v := os.Getenv("CLAWSEC_AI_ENABLED"); v != "" {
		c.AIEnabled, _ = strconv.ParseBool(v)
	}
	if v := os.Getenv("CLAWSEC_AI_ENDPOINT"); v != "" {
		c.AIEndpoint = v
	}
	if v := os.Getenv("CLAWSEC_AI_MODEL"); v != "" {
		c.AIModel = v
	}
	if v := os.Getenv("CLAWSEC_AI_API_KEY"); v != "" {
		c.AIAPIKey = v
	}

	// Product configs: CLAWSEC_SAFELINE_URL, CLAWSEC_SAFELINE_API_KEY, etc.
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimPrefix(parts[0], prefix)
		upperKey := strings.ToUpper(key)

		for _, prod := range []string{"SAFELINE", "XRAY", "CLOUDWALKER", "TANSWER", "DDR"} {
			if strings.HasPrefix(upperKey, prod+"_") {
				field := strings.TrimPrefix(upperKey, prod+"_")
				prodName := strings.ToLower(prod)
				pc := c.Products[prodName]
				switch field {
				case "URL":
					pc.URL = parts[1]
				case "API_KEY":
					pc.APIKey = parts[1]
				case "TOKEN":
					pc.Token = parts[1]
				case "USERNAME":
					pc.Username = parts[1]
				case "PASSWORD":
					pc.Password = parts[1]
				case "INSECURE":
					pc.Insecure, _ = strconv.ParseBool(parts[1])
				}
				c.Products[prodName] = pc
				break
			}
		}
	}
}

// toMap exports Config to a map for YAML serialization.
func (c *Config) toMap() map[string]interface{} {
	m := map[string]interface{}{
		"output_format":   c.OutputFormat,
		"timeout":         c.Timeout,
		"threads":         c.Threads,
		"rate_limit":      c.RateLimit,
		"user_agent":      c.UserAgent,
		"random_ua":       c.RandomUA,
		"proxy":           c.Proxy,
		"force_proxy":     c.ForceProxy,
		"insecure_ssl":    c.InsecureSSL,
		"follow_redirects": c.FollowRedirects,
		"max_redirects":   c.MaxRedirects,
	}

	if c.AIEnabled || c.AIEndpoint != "" || c.AIModel != "" || c.AIAPIKey != "" {
		m["ai"] = map[string]interface{}{
			"enabled":  c.AIEnabled,
			"endpoint": c.AIEndpoint,
			"model":    c.AIModel,
			"api_key":  c.AIAPIKey,
		}
	}

	if len(c.Products) > 0 {
		prods := make(map[string]interface{})
		for name, pc := range c.Products {
			if pc.URL != "" || pc.APIKey != "" {
				prods[name] = map[string]interface{}{
					"url":      pc.URL,
					"api_key":  pc.APIKey,
					"token":    pc.Token,
					"username": pc.Username,
					"password": pc.Password,
					"insecure": pc.Insecure,
				}
			}
		}
		if len(prods) > 0 {
			m["products"] = prods
		}
	}

	return m
}

func toEnvKey(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToUpper(string(result))
}

func setField(fv reflect.Value, v interface{}) {
	switch fv.Kind() {
	case reflect.String:
		if s, ok := v.(string); ok {
			fv.SetString(s)
		}
	case reflect.Int, reflect.Int64:
		switch n := v.(type) {
		case int:
			fv.SetInt(int64(n))
		case int64:
			fv.SetInt(n)
		case float64:
			fv.SetInt(int64(n))
		}
	case reflect.Bool:
		if b, ok := v.(bool); ok {
			fv.SetBool(b)
		}
	case reflect.Slice:
		if arr, ok := v.([]interface{}); ok && fv.Type().Elem().Kind() == reflect.String {
			strs := make([]string, len(arr))
			for i, item := range arr {
				strs[i] = fmt.Sprintf("%v", item)
			}
			fv.Set(reflect.ValueOf(strs))
		}
	}
}

func setFieldFromString(fv reflect.Value, s string) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(s)
	case reflect.Int, reflect.Int64:
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			fv.SetInt(n)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(s); err == nil {
			fv.SetBool(b)
		}
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(s, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			fv.Set(reflect.ValueOf(parts))
		}
	}
}
