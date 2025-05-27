package vsmod

import (
	"github.com/dominikbraun/graph"
)

func GraphFromInfos(infos []*InfoWithFilename, directed bool) (graph.Graph[ID, *InfoWithFilename], error) {
	g := graph.New[ID, *InfoWithFilename](func(i *InfoWithFilename) ID { return i.ModID }, graph.Directed())

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
			if !directed {
				if err := g.AddEdge(depID, info.ModID); err != nil {
					return nil, err
				}
			}
		}
	}

	return g, nil
}

func SortedComponents(infos []*InfoWithFilename) ([][]*InfoWithFilename, error) {
	infoByID := make(map[ID]*InfoWithFilename, len(infos))
	for _, info := range infos {
		infoByID[info.ModID] = info
	}

	g, err := GraphFromInfos(infos, false)
	if err != nil {
		return nil, err
	}

	compIDs, err := graph.StronglyConnectedComponents(g)
	if err != nil {
		return nil, err
	}

	var components [][]*InfoWithFilename

	for _, comp := range compIDs {
		component := make([]*InfoWithFilename, 0, len(comp))
		for _, id := range comp {
			component = append(component, infoByID[id])
		}

		dg, err := GraphFromInfos(component, true)
		if err != nil {
			return nil, err
		}

		order, err := graph.StableTopologicalSort(dg, func(a, b ID) bool { return a < b })
		if err != nil {
			return nil, err
		}

		for i, id := range order {
			component[i] = infoByID[id]
		}

		components = append(components, component)
	}

	return components, nil
}

func BisectComponents(components [][]*InfoWithFilename) ([][]*InfoWithFilename, [][]*InfoWithFilename, error) {
	var left [][]*InfoWithFilename
	var right = [][]*InfoWithFilename{}

	var total int
	for _, comp := range components {
		total += len(comp)
	}

	slices.SortFunc(components, func(a, b []*InfoWithFilename) int {
		return cmp.Compare(len(a), len(b))
	})

	for _, comp := range components {
		if len(left) < len(components)/2 {
			left = append(left, comp)
		} else {
			right = append(right, comp)
		}
	}
	return left, right, nil
}
