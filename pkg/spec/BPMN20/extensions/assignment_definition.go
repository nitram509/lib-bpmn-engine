package extensions

import "strings"

type TAssignmentDefinition struct {
	Assignee        string `xml:"assignee,attr"`
	CandidateGroups string `xml:"candidateGroups,attr"`
}

func (ad TAssignmentDefinition) GetCandidateGroups() []string {
	groups := strings.Split(ad.CandidateGroups, ",")
	for i, group := range groups {
		groups[i] = strings.TrimSpace(group)
	}
	return groups
}
