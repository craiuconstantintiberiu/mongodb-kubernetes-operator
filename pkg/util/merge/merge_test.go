package merge

import (
	"reflect"
	"testing"

	"github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/container"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestMergeStringSlices(t *testing.T) {
	type args struct {
		original []string
		override []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Does not include duplicate entries",
			args: args{
				original: []string{"a", "b", "c"},
				override: []string{"a", "c"},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "Adds elements from override",
			args: args{
				original: []string{"a", "b", "c"},
				override: []string{"a", "b", "c", "d", "e"},
			},
			want: []string{"a", "b", "c", "d", "e"},
		},
		{
			name: "Doesn't panic with nil input",
			args: args{
				original: nil,
				override: nil,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringSlices(tt.args.original, tt.args.override); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeStringSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeContainer(t *testing.T) {
	// TODO: Add additional fields as they are merged

	defaultContainer := container.New(
		container.WithName("default-container"),
		container.WithCommand([]string{"a", "b", "c"}),
		container.WithImage("default-image"),
		container.WithImagePullPolicy(corev1.PullAlways),
		container.WithWorkDir("default-work-dir"),
		container.WithArgs([]string{"arg0", "arg1"}),
	)

	t.Run("Override Fields", func(t *testing.T) {
		overrideContainer := container.New(
			container.WithName("override-container"),
			container.WithCommand([]string{"d", "f", "e"}),
			container.WithImage("override-image"),
			container.WithWorkDir("override-work-dir"),
			container.WithArgs([]string{"arg3", "arg2"}),
		)
		mergedContainer := Container(defaultContainer, overrideContainer)
		assert.Equal(t, overrideContainer.Name, mergedContainer.Name, "Name was overridden, and should be used.")
		assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, mergedContainer.Command, "Command was specified in both, so results should be merged.")
		assert.Equal(t, overrideContainer.Image, mergedContainer.Image, "Image was overridden, and should be used.")
		assert.Equal(t, defaultContainer.ImagePullPolicy, mergedContainer.ImagePullPolicy, "No ImagePullPolicy was specified in the override, so the default should be used.")
		assert.Equal(t, overrideContainer.WorkingDir, mergedContainer.WorkingDir)
		assert.Equal(t, []string{"arg0", "arg1", "arg2", "arg3"}, mergedContainer.Args, "Args were specified in both, so results should be merged.")
	})

	t.Run("No Override Fields", func(t *testing.T) {
		mergedContainer := Container(defaultContainer, corev1.Container{})
		assert.Equal(t, defaultContainer.Name, mergedContainer.Name, "Name was not overridden, and should not be used.")
		assert.Equal(t, defaultContainer.Command, mergedContainer.Command, "Command was not specified. The original Command should be used.")
		assert.Equal(t, defaultContainer.Image, mergedContainer.Image, "Image was not overridden, and should not be used.")
		assert.Equal(t, defaultContainer.ImagePullPolicy, mergedContainer.ImagePullPolicy, "No ImagePullPolicy was specified in the override, so the default should be used.")
		assert.Equal(t, defaultContainer.WorkingDir, mergedContainer.WorkingDir)
		assert.Equal(t, defaultContainer.Args, mergedContainer.Args, "Args were not specified. The original Args should be used.")
	})
}

func TestMergeContainerPort(t *testing.T) {
	original := corev1.ContainerPort{
		Name:          "original-port",
		HostPort:      10,
		ContainerPort: 10,
		Protocol:      corev1.ProtocolTCP,
		HostIP:        "4.3.2.1",
	}

	t.Run("Override Fields", func(t *testing.T) {
		override := corev1.ContainerPort{
			Name:          "override-port",
			HostPort:      1,
			ContainerPort: 5,
			Protocol:      corev1.ProtocolUDP,
			HostIP:        "1.2.3.4",
		}
		mergedPort := ContainerPorts(original, override)

		assert.Equal(t, override.Name, mergedPort.Name)
		assert.Equal(t, override.HostPort, mergedPort.HostPort)
		assert.Equal(t, override.ContainerPort, mergedPort.ContainerPort)
		assert.Equal(t, override.HostIP, mergedPort.HostIP)
		assert.Equal(t, override.ContainerPort, mergedPort.ContainerPort)

	})

	t.Run("No Override Fields", func(t *testing.T) {
		mergedPort := ContainerPorts(original, corev1.ContainerPort{})
		assert.Equal(t, original.Name, mergedPort.Name)
		assert.Equal(t, original.HostPort, mergedPort.HostPort)
		assert.Equal(t, original.ContainerPort, mergedPort.ContainerPort)
		assert.Equal(t, original.HostIP, mergedPort.HostIP)
		assert.Equal(t, original.ContainerPort, mergedPort.ContainerPort)
	})
}

func TestMergeVolumeMount(t *testing.T) {
	hostToContainer := corev1.MountPropagationHostToContainer
	hostToContainerRef := &hostToContainer
	original := corev1.VolumeMount{
		Name:             "override-name",
		ReadOnly:         true,
		MountPath:        "override-mount-path",
		SubPath:          "override-sub-path",
		MountPropagation: hostToContainerRef,
		SubPathExpr:      "override-sub-path-expr",
	}

	t.Run("With Override", func(t *testing.T) {
		bidirectional := corev1.MountPropagationBidirectional
		bidirectionalRef := &bidirectional
		override := corev1.VolumeMount{
			Name:             "override-name",
			ReadOnly:         true,
			MountPath:        "override-mount-path",
			SubPath:          "override-sub-path",
			MountPropagation: bidirectionalRef,
			SubPathExpr:      "override-sub-path-expr",
		}
		mergedVolumeMount := VolumeMount(original, override)

		assert.Equal(t, override.Name, mergedVolumeMount.Name)
		assert.Equal(t, override.ReadOnly, mergedVolumeMount.ReadOnly)
		assert.Equal(t, override.MountPath, mergedVolumeMount.MountPath)
		assert.Equal(t, override.MountPropagation, mergedVolumeMount.MountPropagation)
		assert.Equal(t, override.SubPathExpr, mergedVolumeMount.SubPathExpr)
	})

	t.Run("No Override", func(t *testing.T) {
		mergedVolumeMount := VolumeMount(original, corev1.VolumeMount{})

		assert.Equal(t, original.Name, mergedVolumeMount.Name)
		assert.Equal(t, original.ReadOnly, mergedVolumeMount.ReadOnly)
		assert.Equal(t, original.MountPath, mergedVolumeMount.MountPath)
		assert.Equal(t, original.MountPropagation, mergedVolumeMount.MountPropagation)
		assert.Equal(t, original.SubPathExpr, mergedVolumeMount.SubPathExpr)
	})
}

func TestContainerPortSlicesByName(t *testing.T) {

	original := []corev1.ContainerPort{
		{
			Name:          "original-port-0",
			HostPort:      10,
			ContainerPort: 10,
			Protocol:      corev1.ProtocolTCP,
			HostIP:        "1.2.3.4",
		},
		{
			Name:          "original-port-1",
			HostPort:      20,
			ContainerPort: 20,
			Protocol:      corev1.ProtocolTCP,
			HostIP:        "1.2.3.5",
		},
		{
			Name:          "original-port-2",
			HostPort:      30,
			ContainerPort: 30,
			Protocol:      corev1.ProtocolTCP,
			HostIP:        "1.2.3.6",
		},
	}

	override := []corev1.ContainerPort{
		{
			Name:          "original-port-0",
			HostPort:      50,
			ContainerPort: 50,
			Protocol:      corev1.ProtocolTCP,
			HostIP:        "1.2.3.10",
		},
		{
			Name:          "original-port-1",
			HostPort:      60,
			ContainerPort: 60,
			Protocol:      corev1.ProtocolTCP,
			HostIP:        "1.2.3.50",
		},
		{
			Name:          "original-port-3",
			HostPort:      40,
			ContainerPort: 40,
			Protocol:      corev1.ProtocolTCP,
			HostIP:        "1.2.3.6",
		},
	}

	merged := ContainerPortSlicesByName(original, override)

	assert.Len(t, merged, 4, "There are 4 distinct names between the two slices.")

	t.Run("Test Port 0", func(t *testing.T) {
		assert.Equal(t, "original-port-0", merged[0].Name, "The name should remain unchanged")
		assert.Equal(t, int32(50), merged[0].HostPort, "The HostPort should have been overridden")
		assert.Equal(t, int32(50), merged[0].ContainerPort, "The ContainerPort should have been overridden")
		assert.Equal(t, "1.2.3.10", merged[0].HostIP, "The HostIP should have been overridden")
		assert.Equal(t, corev1.ProtocolTCP, merged[0].Protocol, "The Protocol should remain unchanged")
	})
	t.Run("Test Port 1", func(t *testing.T) {
		assert.Equal(t, "original-port-1", merged[1].Name, "The name should remain unchanged")
		assert.Equal(t, int32(60), merged[1].HostPort, "The HostPort should have been overridden")
		assert.Equal(t, int32(60), merged[1].ContainerPort, "The ContainerPort should have been overridden")
		assert.Equal(t, "1.2.3.50", merged[1].HostIP, "The HostIP should have been overridden")
		assert.Equal(t, corev1.ProtocolTCP, merged[1].Protocol, "The Protocol should remain unchanged")
	})
	t.Run("Test Port 2", func(t *testing.T) {
		assert.Equal(t, "original-port-2", merged[2].Name, "The name should remain unchanged")
		assert.Equal(t, int32(30), merged[2].HostPort, "The HostPort should remain unchanged")
		assert.Equal(t, int32(30), merged[2].ContainerPort, "The ContainerPort should remain unchanged")
		assert.Equal(t, "1.2.3.6", merged[2].HostIP, "The HostIP should remain unchanged")
		assert.Equal(t, corev1.ProtocolTCP, merged[2].Protocol, "The Protocol should remain unchanged")
	})
	t.Run("Test Port 3", func(t *testing.T) {
		assert.Equal(t, "original-port-3", merged[3].Name, "The name should remain unchanged")
		assert.Equal(t, int32(40), merged[3].HostPort, "The HostPort should have been overridden")
		assert.Equal(t, int32(40), merged[3].ContainerPort, "The ContainerPort should have been overridden")
		assert.Equal(t, "1.2.3.6", merged[3].HostIP, "The HostIP should have been overridden")
		assert.Equal(t, corev1.ProtocolTCP, merged[3].Protocol, "The Protocol should remain unchanged")
	})

}