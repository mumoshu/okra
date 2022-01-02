package cell

import okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"

func getWeightAt(totalWeight, numTGs, i int) int {
	var weight int

	if totalWeight > 0 {
		weight = totalWeight / numTGs

		if i == numTGs-1 && numTGs > 1 {
			weight = totalWeight - (weight * (numTGs - 1))
		}
	}

	return weight
}

func redistributeWeights(totalWeight int, desiredTGs []okrav1alpha1.ForwardTargetGroup) map[string]okrav1alpha1.ForwardTargetGroup {
	numTGs := len(desiredTGs)
	result := map[string]okrav1alpha1.ForwardTargetGroup{}

	for i, tg := range desiredTGs {
		result[tg.Name] = okrav1alpha1.ForwardTargetGroup{
			Name:   tg.Name,
			ARN:    tg.ARN,
			Weight: getWeightAt(totalWeight, numTGs, i),
		}
	}

	return result
}

func distributeWeights(totalWeight int, desiredTGs []okrav1alpha1.AWSTargetGroup) map[string]okrav1alpha1.ForwardTargetGroup {
	numTGs := len(desiredTGs)
	result := map[string]okrav1alpha1.ForwardTargetGroup{}

	for i, tg := range desiredTGs {
		result[tg.Name] = okrav1alpha1.ForwardTargetGroup{
			Name:   tg.Name,
			ARN:    tg.Spec.ARN,
			Weight: getWeightAt(totalWeight, numTGs, i),
		}
	}

	return result
}
