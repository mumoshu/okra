package v1alpha1

func init() {
	SchemeBuilder.Register(
		&AnalysisTemplate{},
		&AnalysisTemplateList{},
		&ClusterAnalysisTemplate{},
		&ClusterAnalysisTemplateList{},
		&AnalysisRun{},
		&AnalysisRunList{},
	)
}
