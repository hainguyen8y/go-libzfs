package zfs

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
// TESTS ARE DEPENDED AND MUST RUN IN DEPENDENT ORDER

func Test(t *testing.T) {
	TestPoolCreate(t)
	TestPoolVDevTree(t)
	TestExport(t)
	TestPoolImportSearch(t)
	TestImport(t)
	TestExportForce(t)
	TestImportByGUID(t)
	TestPoolProp(t)
	TestPoolStatusAndState(t)
	TestPoolOpenAll(t)
	TestFailPoolOpen(t)

	TestDatasetCreate(t)
	TestDatasetOpen(t)
	TestDatasetSnapshot(t)
	TestDatasetOpenAll(t)
	TestDatasetSetProperty(t)
	TestDatasetHoldRelease(t)

	TestDatasetDestroy(t)

	TestPoolDestroy(t)

	CleanupVDisks()
}
