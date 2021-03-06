package models

import (
	"fmt"
	"regexp"
	"strconv"
)

type FormationEntry struct {
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Memory int    `json:"memory"`
	Ports  []int  `json:"ports"`
}

type Formation []FormationEntry

func ListFormation(app string) (Formation, error) {
	a, err := GetApp(app)

	if err != nil {
		return nil, err
	}

	release, err := a.LatestRelease()

	if err != nil {
		return nil, err
	}

	if release == nil {
		return Formation{}, nil
	}

	manifest, err := LoadManifest(release.Manifest)

	if err != nil {
		return nil, err
	}

	formation := Formation{}

	for _, me := range manifest {
		count, _ := strconv.Atoi(a.Parameters[fmt.Sprintf("%sDesiredCount", UpperName(me.Name))])
		memory, _ := strconv.Atoi(a.Parameters[fmt.Sprintf("%sMemory", UpperName(me.Name))])

		re := regexp.MustCompile(fmt.Sprintf(`%sPort(\d+)Host`, UpperName(me.Name)))

		ports := []int{}

		for key, _ := range a.Parameters {
			matches := re.FindStringSubmatch(key)

			if len(matches) == 2 {
				port, _ := strconv.Atoi(matches[1])
				ports = append(ports, port)
			}
		}

		formation = append(formation, FormationEntry{
			Name:   me.Name,
			Count:  count,
			Memory: memory,
			Ports:  ports,
		})
	}

	return formation, nil
}

func SetFormation(app, process, count, memory string) error {
	a, err := GetApp(app)

	if err != nil {
		return err
	}

	params := map[string]string{}

	if count != "" {
		params[fmt.Sprintf("%sDesiredCount", UpperName(process))] = count
	}

	if memory != "" {
		params[fmt.Sprintf("%sMemory", UpperName(process))] = memory
	}

	return a.UpdateParams(params)
}
