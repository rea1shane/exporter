package collector

var (
	constLabels map[string]string

	labelList1 = []string{"key_a"}
	labelList2 = []string{"key_a", "key_b"}
)

func init() {
	constLabels = make(map[string]string)
	constLabels["key_c"] = "CONST"
}
