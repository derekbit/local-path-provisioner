package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func loadFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	helperPodYaml, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(helperPodYaml), nil
}

func loadHelperPodFile(helperPodYaml string, allowUnsafe bool) (*v1.Pod, error) {
	helperPodJSON, err := yaml.YAMLToJSON([]byte(helperPodYaml))
	if err != nil {
		return nil, fmt.Errorf("invalid YAMLToJSON the helper pod with helperPodYaml: %v", helperPodYaml)
	}
	p := v1.Pod{}
	err = json.Unmarshal(helperPodJSON, &p)
	if err != nil {
		return nil, fmt.Errorf("invalid unmarshal the helper pod with helperPodJson: %v", string(helperPodJSON))
	}
	if len(p.Spec.Containers) == 0 {
		return nil, fmt.Errorf("helper pod template does not specify any container")
	}
	if err := validateHelperPodTemplate(&p, allowUnsafe); err != nil {
		return nil, err
	}
	return &p, nil
}

func validateHelperPodTemplate(p *v1.Pod, allowUnsafe bool) error {
	if allowUnsafe {
		return nil
	}

	if len(p.Spec.Containers) != 1 {
		return fmt.Errorf("helper pod template must specify exactly one container")
	}
	if len(p.Spec.InitContainers) > 0 {
		return fmt.Errorf("helper pod template must not define initContainers unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if len(p.Spec.EphemeralContainers) > 0 {
		return fmt.Errorf("helper pod template must not define ephemeralContainers unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if len(p.Spec.Volumes) > 0 {
		return fmt.Errorf("helper pod template must not define custom volumes unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if p.Spec.HostNetwork || p.Spec.HostPID || p.Spec.HostIPC {
		return fmt.Errorf("helper pod template must not enable host namespaces unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if p.Spec.NodeName != "" {
		return fmt.Errorf("helper pod template must not set spec.nodeName unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if p.Spec.ServiceAccountName != "" {
		return fmt.Errorf("helper pod template must not set spec.serviceAccountName unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if p.Spec.SecurityContext != nil {
		return fmt.Errorf("helper pod template must not set pod securityContext unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}

	container := p.Spec.Containers[0]
	if container.SecurityContext != nil {
		return fmt.Errorf("helper pod template must not set container securityContext unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if len(container.VolumeMounts) > 0 {
		return fmt.Errorf("helper pod template must not define custom volumeMounts unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if len(container.EnvFrom) > 0 {
		return fmt.Errorf("helper pod template must not define envFrom unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	for _, env := range container.Env {
		if env.ValueFrom != nil {
			return fmt.Errorf("helper pod template must not define env.valueFrom unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
		}
	}
	if container.Lifecycle != nil {
		return fmt.Errorf("helper pod template must not define container lifecycle hooks unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if container.LivenessProbe != nil {
		return fmt.Errorf("helper pod template must not define container livenessProbe unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if container.ReadinessProbe != nil {
		return fmt.Errorf("helper pod template must not define container readinessProbe unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}
	if container.StartupProbe != nil {
		return fmt.Errorf("helper pod template must not define container startupProbe unless %s is enabled", FlagAllowUnsafeHelperPodTemplate)
	}

	return nil
}
