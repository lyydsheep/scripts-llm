package role

import (
	"github.com/gin-gonic/gin"
	"scripts-llm/internal"
)

func Query(ctx *gin.Context) {
	// 获取 script_id 参数
	scriptID := ctx.Param("script_id")
	if scriptID == "" {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "script_id is required",
		})
		return
	}

	// 从数据库中获取角色数据
	roles, err := getRolesByScriptID(scriptID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    1,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	resp := ToVo(roles)

	// 构建响应结构
	response := gin.H{
		"code":    0,
		"status":  "success",
		"message": "查询成功",
		"data": gin.H{
			"roles": resp,
		},
	}

	// 返回响应
	ctx.JSON(200, response)
}

// getRolesByScriptID 根据 script_id 从数据库中获取角色数据
func getRolesByScriptID(scriptID string) ([]internal.Role, error) {
	var roles []internal.Role
	result := internal.DB().Where("sid = ?", scriptID).Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

type Vo struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Character     string `json:"character"`
	LanguageHabit string `json:"language_habit"`
}

func ToVo(roles []internal.Role) []Vo {
	resp := make([]Vo, len(roles))
	for i := range roles {
		resp[i].Id = roles[i].Rid
		resp[i].Name = roles[i].Name
		resp[i].Character = roles[i].Character
		resp[i].LanguageHabit = roles[i].LanguageHabit
	}
	return resp
}

// Update 更新角色信息
func Update(ctx *gin.Context) {
	// 定义请求结构体
	type Request struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		Character     string `json:"character" binding:"required"`
		LanguageHabit string `json:"language_habit" binding:"required"`
	}
	var req Request

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "invalid request body",
		})
		return
	}

	// 根据 id 获取角色
	var role internal.Role
	result := internal.DB().Where("rid = ?", req.Id).First(&role)
	if result.Error != nil {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "role not found",
		})
		return
	}

	// 更新角色信息
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Character != "" {
		role.Character = req.Character
	}
	if req.LanguageHabit != "" {
		role.LanguageHabit = req.LanguageHabit
	}

	// 保存更新
	if err := internal.DB().Save(&role).Error; err != nil {
		ctx.JSON(500, gin.H{
			"code":    1,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 构建响应结构
	response := gin.H{
		"code":    0,
		"status":  "success",
		"message": "更新成功",
	}

	// 返回响应
	ctx.JSON(200, response)
}
