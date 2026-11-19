package university

import (
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "encoding/base64"
 "net/http"
 "strconv"
 "strings"
 "time"

 "github.com/flipped-aurora/gin-vue-admin/server/global"
 "github.com/flipped-aurora/gin-vue-admin/server/middleware"
 "github.com/gin-gonic/gin"
 "golang.org/x/crypto/bcrypt"
 "gorm.io/gorm"
)

// V13 student sessions are intentionally separate from GVA administrator JWTs.
// Never add student credentials to sys_users or assign the administrator role 2001.
type studentAccountRow struct {
 ID uint64 `json:"id" gorm:"column:id"`
 TenantID uint64 `json:"-" gorm:"column:tenant_id"`
 StudentID uint64 `json:"studentId" gorm:"column:student_id"`
 PasswordHash string `json:"-" gorm:"column:password_hash"`
 Status int `json:"status" gorm:"column:status"`
}
func (studentAccountRow) TableName() string { return "uni_student_accounts" }
type studentAccountView struct {
 ID uint64 `json:"id" gorm:"column:id"`
 StudentID uint64 `json:"studentId" gorm:"column:student_id"`
 StudentNo string `json:"studentNo" gorm:"column:student_no"`
 Name string `json:"name" gorm:"column:name"`
 Status int `json:"status" gorm:"column:status"`
 LastLoginAt *time.Time `json:"lastLoginAt" gorm:"column:last_login_at"`
}
func studentAuthError(c *gin.Context, status int, message string) { c.JSON(status,gin.H{"code":7,"msg":message}) }
func studentAuthOK(c *gin.Context, data interface{}) { c.JSON(http.StatusOK,gin.H{"code":0,"msg":"成功","data":data}) }
func studentPasswordValid(s string) bool { return len(s)>=12 && len(s)<=72 && strings.TrimSpace(s)==s }
func studentHashToken(token string) string { sum:=sha256.Sum256([]byte(token));return hex.EncodeToString(sum[:]) }
func studentAccountList(c *gin.Context) {
 tid,ok:=middleware.GetTenantID(c);if !ok {studentAuthError(c,403,"无高校身份");return}
 page,_:=strconv.Atoi(c.DefaultQuery("page","1"));if page<1 {page=1}
 size,_:=strconv.Atoi(c.DefaultQuery("pageSize","20"));if size<1 {size=20};if size>100 {size=100}
 q:=global.GVA_DB.Table("uni_students s").Joins("LEFT JOIN uni_student_accounts a ON a.tenant_id=s.tenant_id AND a.student_id=s.id").Where("s.tenant_id=?",tid)
 if keyword:=strings.TrimSpace(c.Query("keyword"));keyword!="" {like:="%"+keyword+"%";q=q.Where("s.name LIKE ? OR s.student_no LIKE ?",like,like)}
 var total int64;if err:=q.Count(&total).Error;err!=nil {studentAuthError(c,500,"查询失败");return}
 var rows []studentAccountView
 err:=q.Select("COALESCE(a.id,0) id,s.id student_id,s.student_no,s.name,COALESCE(a.status,0) status,a.last_login_at").Order("s.id DESC").Limit(size).Offset((page-1)*size).Scan(&rows).Error
 if err!=nil {studentAuthError(c,500,"查询失败");return}
 studentAuthOK(c,gin.H{"list":rows,"total":total})
}
func studentAccountCreate(c *gin.Context) {
 tid,ok:=middleware.GetTenantID(c);if !ok {studentAuthError(c,403,"无高校身份");return}
 var in struct {StudentID uint64 `json:"studentId"`; Password string `json:"password"`}
 if c.ShouldBindJSON(&in)!=nil||in.StudentID==0||!studentPasswordValid(in.Password) {studentAuthError(c,400,"学生ID或密码不合法（密码至少12位，最多72字节）");return}
 hash,err:=bcrypt.GenerateFromPassword([]byte(in.Password),bcrypt.DefaultCost);if err!=nil {studentAuthError(c,500,"密码处理失败");return}
 err=global.GVA_DB.Transaction(func(tx *gorm.DB)error{
  var count int64
  if e:=tx.Table("uni_students").Where("tenant_id=? AND id=? AND status=1",tid,in.StudentID).Count(&count).Error;e!=nil{return e};if count!=1{return gorm.ErrRecordNotFound}
  return tx.Table("uni_student_accounts").Create(map[string]interface{}{"tenant_id":tid,"student_id":in.StudentID,"password_hash":string(hash),"status":1}).Error
 })
 if err!=nil {studentAuthError(c,400,"学生不存在、已停用或账号已创建");return}
 studentAuthOK(c,gin.H{"created":true})
}
func studentAccountReset(c *gin.Context) {
 tid,ok:=middleware.GetTenantID(c);if !ok {studentAuthError(c,403,"无高校身份");return}
 var in struct {StudentID uint64 `json:"studentId"`; Password string `json:"password"`}
 if c.ShouldBindJSON(&in)!=nil||in.StudentID==0||!studentPasswordValid(in.Password) {studentAuthError(c,400,"参数不合法");return}
 hash,err:=bcrypt.GenerateFromPassword([]byte(in.Password),bcrypt.DefaultCost);if err!=nil {studentAuthError(c,500,"密码处理失败");return}
 err=global.GVA_DB.Transaction(func(tx *gorm.DB)error{
  r:=tx.Table("uni_student_accounts").Where("tenant_id=? AND student_id=?",tid,in.StudentID).Updates(map[string]interface{}{"password_hash":string(hash),"password_changed_at":time.Now()})
  if r.Error!=nil{return r.Error};if r.RowsAffected!=1{return gorm.ErrRecordNotFound}
  return tx.Exec("UPDATE uni_student_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=(SELECT id FROM uni_student_accounts WHERE tenant_id=? AND student_id=?) AND revoked_at IS NULL",tid,tid,in.StudentID).Error
 })
 if err!=nil {studentAuthError(c,404,"账号不存在");return};studentAuthOK(c,gin.H{"reset":true})
}
func studentAccountStatus(c *gin.Context) {
 tid,ok:=middleware.GetTenantID(c);if !ok {studentAuthError(c,403,"无高校身份");return}
 var in struct {StudentID uint64 `json:"studentId"`;Status int `json:"status"`}
 if c.ShouldBindJSON(&in)!=nil||in.StudentID==0||(in.Status!=1&&in.Status!=2) {studentAuthError(c,400,"参数不合法");return}
 err:=global.GVA_DB.Transaction(func(tx *gorm.DB)error{
  r:=tx.Table("uni_student_accounts").Where("tenant_id=? AND student_id=?",tid,in.StudentID).Update("status",in.Status)
  if r.Error!=nil{return r.Error};if r.RowsAffected!=1{return gorm.ErrRecordNotFound}
  if in.Status==2{return tx.Exec("UPDATE uni_student_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=(SELECT id FROM uni_student_accounts WHERE tenant_id=? AND student_id=?) AND revoked_at IS NULL",tid,tid,in.StudentID).Error}
  return nil
 })
 if err!=nil {studentAuthError(c,404,"账号不存在");return};studentAuthOK(c,gin.H{"updated":true})
}
func studentPortalLogin(c *gin.Context) {
 var in struct {TenantID uint64 `json:"tenantId"`; StudentNo string `json:"studentNo"`;Password string `json:"password"`}
 if c.ShouldBindJSON(&in)!=nil||in.TenantID==0||len(in.StudentNo)==0||len(in.StudentNo)>64||len(in.Password)>256 {studentAuthError(c,400,"参数不合法");return}
 var row struct {ID uint64 `gorm:"column:id"`;TenantID uint64 `gorm:"column:tenant_id"`;StudentID uint64 `gorm:"column:student_id"`;PasswordHash string `gorm:"column:password_hash"`}
 err:=global.GVA_DB.Table("uni_student_accounts a").Select("a.id,a.tenant_id,a.student_id,a.password_hash").
  Joins("JOIN uni_students s ON s.tenant_id=a.tenant_id AND s.id=a.student_id").
  Joins("JOIN uni_tenants t ON t.id=a.tenant_id").
  Where("a.tenant_id=? AND s.student_no=? AND a.status=1 AND s.status=1 AND t.status=1",in.TenantID,strings.TrimSpace(in.StudentNo)).Take(&row).Error
 if err!=nil||bcrypt.CompareHashAndPassword([]byte(row.PasswordHash),[]byte(in.Password))!=nil {studentAuthError(c,401,"高校、学号或密码错误");return}
 random:=make([]byte,32);if _,err=rand.Read(random);err!=nil {studentAuthError(c,500,"创建会话失败");return}
 token:=base64.RawURLEncoding.EncodeToString(random)
 expires:=time.Now().Add(12*time.Hour)
 err=global.GVA_DB.Transaction(func(tx *gorm.DB)error{
  if e:=tx.Table("uni_student_sessions").Create(map[string]interface{}{"tenant_id":row.TenantID,"account_id":row.ID,"token_hash":studentHashToken(token),"expires_at":expires}).Error;e!=nil{return e}
  return tx.Table("uni_student_accounts").Where("id=? AND tenant_id=?",row.ID,row.TenantID).Update("last_login_at",time.Now()).Error
 })
 if err!=nil {studentAuthError(c,500,"创建会话失败");return}
 studentAuthOK(c,gin.H{"token":token,"expiresAt":expires,"studentId":row.StudentID})
}
const studentPortalStudentID="uniPortalStudentID"
const studentPortalTenantID="uniPortalTenantID"
const studentPortalAccountID="uniPortalAccountID"
func studentPortalAuth() gin.HandlerFunc {return func(c *gin.Context){
 auth:=c.GetHeader("Authorization")
 if !strings.HasPrefix(auth,"Bearer ") {studentAuthError(c,401,"请先登录学生端");c.Abort();return}
 token:=strings.TrimPrefix(auth,"Bearer ")
 if len(token)!=43 {studentAuthError(c,401,"登录凭证无效");c.Abort();return}
 var row struct {AccountID uint64 `gorm:"column:account_id"`;StudentID uint64 `gorm:"column:student_id"`;TenantID uint64 `gorm:"column:tenant_id"`}
 err:=global.GVA_DB.Table("uni_student_sessions sess").Select("sess.account_id,a.student_id,a.tenant_id").
  Joins("JOIN uni_student_accounts a ON a.tenant_id=sess.tenant_id AND a.id=sess.account_id").
  Joins("JOIN uni_students s ON s.tenant_id=a.tenant_id AND s.id=a.student_id").
  Joins("JOIN uni_tenants t ON t.id=a.tenant_id").
  Where("sess.token_hash=? AND sess.revoked_at IS NULL AND sess.expires_at>? AND a.status=1 AND s.status=1 AND t.status=1",studentHashToken(token),time.Now()).Take(&row).Error
 if err!=nil {studentAuthError(c,401,"登录已失效，请重新登录");c.Abort();return}
 c.Set(studentPortalStudentID,row.StudentID);c.Set(studentPortalTenantID,row.TenantID);c.Set(studentPortalAccountID,row.AccountID);c.Next()
}}
func studentPortalMe(c *gin.Context) {
 tid:=c.GetUint64(studentPortalTenantID);sid:=c.GetUint64(studentPortalStudentID)
 var row struct {StudentNo string `json:"studentNo"`;Name string `json:"name"`;ClassID uint64 `json:"classId"`;EnrollmentYear int `json:"enrollmentYear"`}
 err:=global.GVA_DB.Table("uni_students").Select("student_no,name,class_id,enrollment_year").Where("tenant_id=? AND id=?",tid,sid).Take(&row).Error
 if err!=nil {studentAuthError(c,404,"学生档案不存在");return}
 studentAuthOK(c,row)
}
func studentPortalChangePassword(c *gin.Context) {
 aid:=c.GetUint64(studentPortalAccountID);tid:=c.GetUint64(studentPortalTenantID)
 var in struct {OldPassword string `json:"oldPassword"`;NewPassword string `json:"newPassword"`}
 if c.ShouldBindJSON(&in)!=nil||!studentPasswordValid(in.NewPassword) {studentAuthError(c,400,"新密码至少12位，最多72字节");return}
 var account studentAccountRow
 if err:=global.GVA_DB.Where("tenant_id=? AND id=?",tid,aid).Take(&account).Error;err!=nil {studentAuthError(c,401,"账号不存在");return}
 if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash),[]byte(in.OldPassword))!=nil {studentAuthError(c,401,"旧密码错误");return}
 hash,err:=bcrypt.GenerateFromPassword([]byte(in.NewPassword),bcrypt.DefaultCost);if err!=nil {studentAuthError(c,500,"密码处理失败");return}
 err=global.GVA_DB.Transaction(func(tx *gorm.DB)error{
  r:=tx.Table("uni_student_accounts").Where("tenant_id=? AND id=? AND password_hash=?",tid,aid,account.PasswordHash).Updates(map[string]interface{}{"password_hash":string(hash),"password_changed_at":time.Now()})
  if r.Error!=nil{return r.Error};if r.RowsAffected!=1{return gorm.ErrRecordNotFound}
  return tx.Exec("UPDATE uni_student_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=? AND revoked_at IS NULL",tid,aid).Error
 })
 if err!=nil {studentAuthError(c,409,"密码已变化，请重新登录");return}
 studentAuthOK(c,gin.H{"changed":true})
}
func studentPortalLogout(c *gin.Context) {
 token:=strings.TrimPrefix(c.GetHeader("Authorization"),"Bearer ")
 err:=global.GVA_DB.Table("uni_student_sessions").Where("token_hash=? AND tenant_id=? AND account_id=?",studentHashToken(token),c.GetUint64(studentPortalTenantID),c.GetUint64(studentPortalAccountID)).Update("revoked_at",time.Now()).Error
 if err!=nil {studentAuthError(c,500,"退出失败");return};studentAuthOK(c,gin.H{"loggedOut":true})
}
// Register under GVA's PRIVATE group (JWT + Casbin); student sessions cannot call these.
func InitStudentAccountRouter(privateGroup *gin.RouterGroup) {
 g:=privateGroup.Group("uni").Use(middleware.TenantAuth())
 g.GET("/student-accounts",studentAccountList)
 g.POST("/student-accounts",studentAccountCreate)
 g.POST("/student-accounts/reset-password",studentAccountReset)
 g.PUT("/student-accounts/status",studentAccountStatus)
}
// Register under GVA's PUBLIC group, not the JWT-protected private group.
func InitStudentPortalRouter(publicGroup *gin.RouterGroup) {
 g:=publicGroup.Group("uni/student-portal")
 g.POST("/login",studentPortalLogin)
 protected:=g.Group("").Use(studentPortalAuth())
 protected.GET("/me",studentPortalMe)
 protected.POST("/change-password",studentPortalChangePassword)
 protected.POST("/logout",studentPortalLogout)
}
