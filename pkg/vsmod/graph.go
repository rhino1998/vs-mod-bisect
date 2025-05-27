package vsmod

import (
	"slices"

	"github.com/dominikbraun/graph"
)

func GraphFromInfos(infos map[string]*Info) (graph.Graph[ID, *Info], error) {
	g := graph.New[ID, *Info](func(i *Info) ID { return i.ModID }, graph.Directed())

	for _, info := range infos {
		if err := g.AddVertex(info); err != nil {
			return nil, err
		}
	}
	for _, info := range infos {
		for depID := range info.Dependencies {
			switch depID {
			case "game", "survival", "creative", "vanilla":
				continue
			}

			if err := g.AddEdge(info.ModID, depID); err != nil {
				return nil, err
			}
		}
	}

	return g, nil
}

func Bisect(infos map[string]*Info) (map[string]*Info, map[string]*Info, error) {
	infoByID := make(map[ID]*Info, len(infos))
	pathByID := make(map[ID]string, len(infos))
	for path, info := range infos {
		infoByID[info.ModID] = info
		pathByID[info.ModID] = path
	}

	g, err := GraphFromInfos(infos)
	if err != nil {
		return nil, nil, err
	}

	ids, err := graph.StableTopologicalSort(g, func(a, b ID) bool {
		return a < b
	})
	slices.Reverse(ids)
	if err != nil {
		return nil, nil, err
	}

	midPoint := len(infos) / 2

	left := make(map[string]*Info)
	right := make(map[string]*Info)
	for _, id := range ids {
		if len(left) < midPoint {
			left[pathByID[id]] = infoByID[id]
		} else {
			right[pathByID[id]] = infoByID[id]
		}
	}

	return left, right, nil
}
