package utils

// utilPage 分页工具（hutool PageUtil）。
//
// 注意与 Math.PageCount 的关系：TotalPage 直接复用 Math.PageCount 实现，语义一致
// （hutool 中 PageUtil.totalPage 与 NumberUtil.count 亦为同一逻辑），不再重复实现。
type utilPage struct{}

// TransToStartEnd 把页码与每页大小转换为数据库查询区间 [start, end)（前闭后开）。
// TransToStartEnd(2, 10) = [10, 20)，即第 2 页对应 LIMIT 10 OFFSET 10 的行区间。
// 页码从 1 开始（页码 <= 0 时按第 1 页处理）。
// 场景：手工拼 LIMIT/OFFSET 前先算行号区间。
func (p *utilPage) TransToStartEnd(pageNo, pageSize int64) (start, end int64) {
	if pageNo <= 0 {
		pageNo = 1
	}
	start = (pageNo - 1) * pageSize
	end = start + pageSize
	return start, end
}

// GetStart 返回当前页起始行号（0 基，等价于 SQL 的 OFFSET 值）。
// GetStart(3, 10) = 20。
func (p *utilPage) GetStart(pageNo, pageSize int64) int64 {
	if pageNo <= 0 {
		pageNo = 1
	}
	return (pageNo - 1) * pageSize
}

// GetEnd 返回当前页结束行号（0 基、开区间）。GetEnd(3, 10) = 30。
func (p *utilPage) GetEnd(pageNo, pageSize int64) int64 {
	_, end := p.TransToStartEnd(pageNo, pageSize)
	return end
}

// TotalPage 计算总页数（向上取整，复用 Math.PageCount）。
// Page.TotalPage(101, 20) = 6。
func (p *utilPage) TotalPage(totalCount, pageSize int64) int64 {
	return Math.PageCount(totalCount, pageSize)
}

// Rainbow 生成"彩虹式"分页导航页码数组：以当前页为中心、最多 displayCount 个页码。
// Rainbow(5, 20, 7) = [2,3,4,5,6,7,8]。场景：Web 分页条渲染页码按钮。
// 边界处理：总数不足 displayCount 时返回全部页码；靠左/靠右时窗口自动贴边。
func (p *utilPage) Rainbow(pageNo, totalPage, displayCount int64) []int64 {
	if totalPage <= 0 || displayCount <= 0 {
		return nil
	}
	if pageNo <= 0 {
		pageNo = 1
	}
	if displayCount > totalPage {
		displayCount = totalPage
	}
	start := pageNo - displayCount/2
	if start < 1 {
		start = 1
	}
	if start+displayCount-1 > totalPage {
		start = totalPage - displayCount + 1
	}
	result := make([]int64, 0, displayCount)
	for i := start; i < start+displayCount; i++ {
		result = append(result, i)
	}
	return result
}
