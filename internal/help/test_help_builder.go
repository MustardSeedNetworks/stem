package help

type testHelpMeta struct {
	ID       string
	Name     string
	Standard string
	Category string
}

type testHelpDescriptions struct {
	Summary    string
	TechDesc   string
	LaymanDesc string
}

type testHelpUsage struct {
	WhenToUse    string
	WhenNotToUse string
}

type testHelpDetails struct {
	Parameters         []Parameter
	Metrics            []Metric
	SuccessCriteria    string
	FailureExplanation string
	Examples           []Example
	Tips               []string
	CommonIssues       []Issue
	RFCSection         string
	SeeAlso            []string
}

func buildTestHelp(
	meta testHelpMeta,
	desc testHelpDescriptions,
	usage testHelpUsage,
	details testHelpDetails,
) TestHelp {
	return TestHelp{
		ID:                 meta.ID,
		Name:               meta.Name,
		Standard:           meta.Standard,
		Category:           meta.Category,
		Summary:            desc.Summary,
		TechDesc:           desc.TechDesc,
		LaymanDesc:         desc.LaymanDesc,
		WhenToUse:          usage.WhenToUse,
		WhenNotToUse:       usage.WhenNotToUse,
		Parameters:         details.Parameters,
		Metrics:            details.Metrics,
		SuccessCriteria:    details.SuccessCriteria,
		FailureExplanation: details.FailureExplanation,
		Examples:           details.Examples,
		Tips:               details.Tips,
		CommonIssues:       details.CommonIssues,
		RFCSection:         details.RFCSection,
		SeeAlso:            details.SeeAlso,
	}
}
