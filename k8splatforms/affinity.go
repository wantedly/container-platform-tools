package k8splatforms

import (
	"sort"
	"strings"

	"github.com/wantedly/container-platform-tools/dockerplatforms"
	v1 "k8s.io/api/core/v1"
)

type PlatformAffinity2 struct {
	any        bool
	anyByOS    map[string]bool
	anyByArch  map[string]bool
	byPlatform map[string]bool
}

func EmptyPlatformAffinity() PlatformAffinity2 {
	return PlatformAffinity2{
		any:        false,
		anyByOS:    map[string]bool{},
		anyByArch:  map[string]bool{},
		byPlatform: map[string]bool{},
	}
}
func FullPlatformAffinity() PlatformAffinity2 {
	return PlatformAffinity2{
		any:        true,
		anyByOS:    map[string]bool{},
		anyByArch:  map[string]bool{},
		byPlatform: map[string]bool{},
	}
}
func (a *PlatformAffinity2) Matches(os string, arch string) bool {
	if matches, ok := a.byPlatform[os+"/"+arch]; ok {
		return matches
	}
	if matches, ok := a.anyByOS[os]; ok {
		return matches
	}
	if matches, ok := a.anyByArch[arch]; ok {
		return matches
	}
	return a.any
}
func (a *PlatformAffinity2) Normalize() {
	for os, matches := range a.anyByOS {
		if matches == a.any {
			delete(a.anyByOS, os)
		}
	}
	for arch, matches := range a.anyByArch {
		if matches == a.any {
			delete(a.anyByArch, arch)
		}
	}
	for platform, matches := range a.byPlatform {
		os, arch := splitPlatform(platform)
		defaultMatch := a.any
		if matchesByOS, ok := a.anyByOS[os]; ok {
			defaultMatch = matchesByOS
		} else if matchesByArch, ok := a.anyByArch[arch]; ok {
			defaultMatch = matchesByArch
		}
		if matches == defaultMatch {
			delete(a.byPlatform, platform)
		}
	}
}

func splitPlatform(s string) (string, string) {
	idx := strings.IndexRune(s, '/')
	if idx < 0 {
		return s, ""
	}
	return s[:idx], s[idx+1:]
}

func (a *PlatformAffinity2) SetUnion(other PlatformAffinity2) {
}

const Wildcard = "*"

type PlatformAffinity []dockerplatforms.DockerPlatform

func (a PlatformAffinity) Normalize() PlatformAffinity {
	result := make(PlatformAffinity, len(a))
	copy(result, a)
	sort.Slice(result, func(i, j int) bool {
		lhs := result[i]
		rhs := result[j]
		lhsRank := sortRank(lhs)
		rhsRank := sortRank(rhs)
		if lhsRank != rhsRank {
			return lhsRank < rhsRank
		} else if lhs.OS != rhs.OS {
			return lhs.OS < rhs.OS
		} else {
			return lhs.Architecture < rhs.Architecture
		}
	})
	newLen := 0
	// Quadratic but small enough
	for i := 0; i < len(result); i++ {
		needed := true
		for j := 0; j < newLen; j++ {
			if incl(result[i].OS, result[j].OS) && incl(result[i].Architecture, result[j].Architecture) {
				needed = false
				break
			}
		}
		if needed {
			result[newLen] = result[i]
			newLen++
		}
	}
	return result[:newLen]
}

func incl(lhs, rhs string) bool {
	return rhs == Wildcard || lhs == rhs
}

func sortRank(cond dockerplatforms.DockerPlatform) int {
	if cond.OS == Wildcard && cond.Architecture == Wildcard {
		return 0
	} else if cond.OS == Wildcard {
		return 1
	} else if cond.Architecture == Wildcard {
		return 2
	} else {
		return 3
	}
}

func (a PlatformAffinity) Union(other PlatformAffinity) PlatformAffinity {
	result := PlatformAffinity(append(a, other...))
	return result.Normalize()
}

func (a PlatformAffinity) Intersect(other PlatformAffinity) PlatformAffinity {
	result := PlatformAffinity{}
	for _, lhs := range a {
		for _, rhs := range other {
			os := lhs.OS
			arch := lhs.Architecture
			if os == Wildcard {
				os = rhs.OS
			} else if rhs.OS != Wildcard && os != rhs.OS {
				continue
			}
			if arch == Wildcard {
				arch = rhs.Architecture
			} else if rhs.Architecture != Wildcard && arch != rhs.Architecture {
				continue
			}
			result = append(result, dockerplatforms.DockerPlatform{OS: os, Architecture: arch})
		}
	}
	return result.Normalize()
}

func GetPodPlatformAffinity(pod *v1.Pod) PlatformAffinity {
	var nodeSelector *v1.NodeSelector
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			nodeSelector = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}
	return GetPlatformAffinity(pod.Spec.NodeSelector, nodeSelector)
}

func GetPlatformAffinity(simpleNodeSelector map[string]string, nodeSelector *v1.NodeSelector) PlatformAffinity {
	// Based on:
	// - PreFilter https://github.com/kubernetes/kubernetes/blob/v1.30.2/pkg/scheduler/framework/plugins/nodeaffinity/node_affinity.go
	// - Match https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go

	platformsFromNodeSelector := PlatformAffinity{}
	if nodeSelector != nil {
		for _, term := range nodeSelector.NodeSelectorTerms {
			anded := PlatformAffinity{dockerplatforms.DockerPlatform{OS: Wildcard, Architecture: Wildcard}}
			for _, expr := range term.MatchExpressions {
				if expr.Key == "kubernetes.io/arch" || expr.Key == "beta.kubernetes.io/arch" {
					switch expr.Operator {
					case v1.NodeSelectorOpIn:
					}
				}
			}
		}
	}
	panic("not implemented")
}
