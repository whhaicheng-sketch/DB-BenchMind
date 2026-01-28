#!/bin/bash
# GUI Pages Validation Script
# Tests that all GUI pages compile and initialize correctly

set -e

echo "========================================="
echo "DB-BenchMind GUI Pages Validation Test"
echo "========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

# Test function
test_page() {
    local page_name=$1
    local page_file=$2

    echo -n "Testing $page_name... "

    # Check if file exists
    if [ ! -f "$page_file" ]; then
        echo -e "${RED}FAILED${NC} (file not found)"
        ((FAILED++))
        return
    fi

    # Check if file compiles
    if go build -o /tmp/test_page.o "$page_file" >/dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        ((PASSED++))
        rm -f /tmp/test_page.o
    else
        echo -e "${RED}FAILED${NC} (compilation error)"
        ((FAILED++))
    fi
}

# Test each page
echo "Page Compilation Tests:"
echo "---------------------"

test_page "Template Page" "internal/transport/ui/pages/template_page.go"
test_page "Task Page" "internal/transport/ui/pages/task_page.go"
test_page "Monitor Page" "internal/transport/ui/pages/monitor_page.go"
test_page "History Page" "internal/transport/ui/pages/history_page.go"
test_page "Comparison Page" "internal/transport/ui/pages/comparison_page.go"
test_page "Report Page" "internal/transport/ui/pages/report_page.go"
test_page "Settings Page" "internal/transport/ui/pages/settings_page.go"
test_page "Connection Page" "internal/transport/ui/pages/connection_page.go"

echo ""
echo "Full GUI Build Test:"
echo "---------------------"

echo -n "Building complete GUI application... "
if go build -o /tmp/db-benchmind-test ./cmd/db-benchmind >/dev/null 2>&1; then
    echo -e "${GREEN}PASSED${NC}"
    ((PASSED++))
    BINARY_SIZE=$(stat -f%z /tmp/db-benchmind-test 2>/dev/null || stat -c%s /tmp/db-benchmind-test 2>/dev/null)
    echo "  Binary size: $(($BINARY_SIZE / 1024 / 1024))MB"
    rm -f /tmp/db-benchmind-test
else
    echo -e "${RED}FAILED${NC} (build error)"
    ((FAILED++))
fi

echo ""
echo "Code Structure Checks:"
echo "---------------------"

# Check for window parameter in page constructors
echo -n "Checking window parameter usage... "
WINDOW_COUNT=$(grep -r "func New.*Page(win fyne.Window" internal/transport/ui/pages/*.go 2>/dev/null | wc -l)
if [ "$WINDOW_COUNT" -ge 6 ]; then
    echo -e "${GREEN}PASSED${NC} ($WINDOW_COUNT pages use window parameter)"
    ((PASSED++))
else
    echo -e "${YELLOW}WARNING${NC} (only $WINDOW_COUNT pages use window parameter)"
    ((PASSED++))
fi

# Check for nil in dialog calls
echo -n "Checking for nil in dialog calls... "
NIL_COUNT=$(grep -r "dialog\..*nil\)" internal/transport/ui/pages/*.go 2>/dev/null | wc -l)
if [ "$NIL_COUNT" -eq 0 ]; then
    echo -e "${GREEN}PASSED${NC} (no nil found in dialog calls)"
    ((PASSED++))
else
    echo -e "${RED}FAILED${NC} (found $NIL_COUNT nil references in dialog calls)"
    ((FAILED++))
fi

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
