package http

import (
	"net/http"

	"enfok_bd/src/domain/node"
	"enfok_bd/src/domain/ports/driving"

	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	svc driving.NodeServicePort
}

func NewNodeHandler(s driving.NodeServicePort) *NodeHandler {
	return &NodeHandler{svc: s}
}

func (h *NodeHandler) Register(c *gin.Context) {
	var n node.Node
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RegisterNode(c.Request.Context(), &n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Node registered successfully"})
}

func (h *NodeHandler) Heartbeat(c *gin.Context) {
	if err := h.svc.Heartbeat(c.Request.Context(), c.Param("node_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat received"})
}

func (h *NodeHandler) List(c *gin.Context) {
	nodes, err := h.svc.ListNodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}
