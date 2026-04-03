package controller

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetTokenOpenClawModels returns the list of models available to a token,
// enriched with resolved capability metadata in OpenClaw-compatible format.
func GetTokenOpenClawModels(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid token id")
		return
	}

	token, err := model.GetTokenById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var modelNames []string
	if token.ModelLimitsEnabled && token.ModelLimits != "" {
		for _, name := range strings.Split(token.ModelLimits, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				modelNames = append(modelNames, name)
			}
		}
	} else {
		modelNames = model.GetEnabledModels()
	}

	models := model.GetModelsWithCapabilities(modelNames)
	common.ApiSuccess(c, models)
}
