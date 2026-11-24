package university

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
	"unicode"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxLeaveProofSize = 5 << 20
const maxLeaveProofCount = 5
const leaveProofColumns = "EXISTS (SELECT 1 FROM uni_leave_proofs p WHERE p.leave_id = l.id) AS has_proof, " +
	"(SELECT COUNT(*) FROM uni_leave_proofs p WHERE p.leave_id = l.id) AS proof_count"

type uniLeaveProof struct {
	LeaveID     uint64 `gorm:"primaryKey;autoIncrement:false"`
	ProofIndex  uint8  `gorm:"primaryKey;autoIncrement:false"`
	Filename    string
	ContentType string
	Content     []byte
}

func (uniLeaveProof) TableName() string { return "uni_leave_proofs" }

func readLeaveProof(header *multipart.FileHeader) (*uniLeaveProof, error) {
	if header.Size > maxLeaveProofSize {
		return nil, errors.New("证明文件不能超过 5 MB")
	}
	file, err := header.Open()
	if err != nil {
		return nil, errors.New("无法读取证明文件")
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxLeaveProofSize+1))
	if err != nil || len(content) == 0 {
		return nil, errors.New("证明文件为空或无法读取")
	}
	if len(content) > maxLeaveProofSize {
		return nil, errors.New("证明文件不能超过 5 MB")
	}

	filename := path.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
	contentType := http.DetectContentType(content)
	extension := strings.ToLower(path.Ext(filename))
	valid := contentType == "application/pdf" && extension == ".pdf" ||
		contentType == "image/png" && extension == ".png" ||
		contentType == "image/jpeg" && (extension == ".jpg" || extension == ".jpeg")
	if !valid || len(filename) > 255 || strings.ContainsFunc(filename, unicode.IsControl) {
		return nil, errors.New("请选择文件名有效的 PDF、JPG 或 PNG 文件")
	}
	return &uniLeaveProof{Filename: filename, ContentType: contentType, Content: content}, nil
}

func downloadLeaveProof(c *gin.Context, scope func(*gorm.DB) *gorm.DB) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		studentAuthError(c, http.StatusBadRequest, "请假单ID不合法")
		return
	}

	index, err := strconv.Atoi(c.DefaultQuery("index", "1"))
	if err != nil || index < 1 || index > maxLeaveProofCount {
		studentAuthError(c, http.StatusBadRequest, "证明文件序号不合法")
		return
	}
	query := global.GVA_DB.Table("uni_leave_proofs p").
		Select("p.filename,p.content_type,p.content").
		Joins("JOIN uni_leave_requests l ON l.id=p.leave_id").
		Where("p.leave_id=? AND p.proof_index=?", id, index)
	var proof uniLeaveProof
	err = scope(query).Take(&proof).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		studentAuthError(c, http.StatusNotFound, "证明文件不存在或无权访问")
		return
	}
	if err != nil {
		studentAuthError(c, http.StatusInternalServerError, "下载证明文件失败")
		return
	}

	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": proof.Filename}))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, proof.ContentType, proof.Content)
}
