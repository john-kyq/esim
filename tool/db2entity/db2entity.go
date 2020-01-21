package db2entity

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os/exec"

	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file_dir"
	logger "github.com/jukylin/esim/log"
	"os"
	"strings"
)

var (
	log logger.Logger
)

func init() {
	log = logger.NewLogger()
}

func GenEntity(v *viper.Viper) error {

	debug := v.GetBool("debug")

	if debug == false {

		if v.GetBool("disrepotar") == false {
			//repo 目录是存在
			existsRepo, err := file_dir.IsExistsDir("./internal/infra/repo")
			if err != nil {
				log.Fatalf(err.Error())
			}

			if existsRepo == false {
				log.Fatalf("dir repo not exists")
			}
		}

		if v.GetBool("disdaotar") == false {
			//dao 目录是否存在
			existsdao, err := file_dir.IsExistsDir("./internal/infra/repo/dao")
			if err != nil {
				log.Fatalf(err.Error())
			}

			if existsdao == false {
				log.Fatalf("dir dao not exists")
			}
		}

		if v.GetBool("disetar") == false {
			//entity目录是存在
			existsEntity, err := file_dir.IsExistsDir("./internal/domain/entity")
			if err != nil {
				log.Fatalf(err.Error())
			}

			if existsEntity == false {
				log.Fatalf("dir entity not exists")
			}
		}
	}

	host := v.GetString("host")
	if host == "" {
		log.Fatalf("host is empty")
	}

	port := v.GetInt("port")
	if port == 0 {
		log.Fatalf("port is 0")
	}

	user := v.GetString("user")
	if user == "" {
		log.Fatalf("user is empty")
	}

	password := v.GetString("password")
	if password == "" {
		log.Fatalf("password is empty")
	}

	table := v.GetString("table")
	if table == "" {
		log.Fatalf("table is empty")
	}

	database := v.GetString("database")
	if host == "" {
		log.Fatalf("database is empty")
	}

	var daoTarget string
	var etar string
	var entityDir string
	var repoTarget string

	if v.GetBool("disetar") == false {

		etar = v.GetString("etar")

		if etar == "" {
			etar = "internal/domain/entity"
		}

		entityDir = etar + "/"
		etar = entityDir + table + ".go"
		ex, err := file_dir.IsExistsFile(etar)
		if err != nil {
			log.Fatalf(err.Error())
		}

		if ex {
			log.Fatalf(etar + " exists")
		}

		log.Infof("create dir ... %s", etar)

		err = file_dir.CreateDir(entityDir)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	if v.GetBool("disdaotar") == false {

		daotar := v.GetString("daotar")
		if daotar == "" {
			daotar = "internal/infra/repo/dao"
		}

		daoDir := daotar + "/"
		daoTarget = daoDir + table + ".go"
		ex, err := file_dir.IsExistsFile(daoTarget)
		if err != nil {
			log.Fatalf(err.Error())
		}

		if ex {
			log.Fatalf(daoDir + " exists")
		}

		log.Infof("create dir ... %s", daoDir)

		err = file_dir.CreateDir(daoDir)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	if v.GetBool("disrepotar") == false {

		repotar := v.GetString("repotar")
		if repotar == "" {
			repotar = "internal/infra/repo"
		}

		repoTarget = repotar + "/" + table + ".go"
		ex, err := file_dir.IsExistsFile(daoTarget)
		if err != nil {
			log.Fatalf(err.Error())
		}

		if ex {
			log.Fatalf(repotar + " exists")
		}

		log.Infof("create dir ... %s", repotar)

		err = file_dir.CreateDir(repotar)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	columnDataTypes, err := GetColumnsFromMysqlTable(user, password, host, port, database, table)

	if err != nil {
		log.Fatalf(err.Error())
	}

	pk := v.GetString("package")
	if pk == "" {
		pk = database
	}

	st := v.GetString("struct")
	if st == "" {
		st = table
	}

	st = camelString(st)

	js := v.GetBool("json")

	gorm := v.GetBool("gorm")

	guregu := v.GetBool("guregu")

	struc, genMysqlInfo, err := Generate(columnDataTypes, table,
		st, pk, js, gorm, guregu, v)
	if err != nil {
		log.Fatalf("Error in creating struct from json: " + err.Error())
	}

	if v.GetBool("disetar") == false {
		err = ioutil.WriteFile(etar, struc, 0666)
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Infof("create file  %s success", etar)
	} else {
		log.Infof("not create file  %s", etar)
	}

	if v.GetBool("disdaotar") == false {

		daoStr, err := GenerateDao(table, st, pk, v, genMysqlInfo)
		if err != nil {
			log.Fatalf("Error in creating struct from json: " + err.Error())
		}

		forDaoStr, err := format.Source([]byte(daoStr))
		if err != nil {
			log.Errorf(err.Error())
		} else {
			err = ioutil.WriteFile(daoTarget, forDaoStr, 0666)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
		log.Infof("create file  %s success", daoTarget)
	} else {
		log.Infof("not create file  %s", daoTarget)
	}

	if v.GetBool("disrepotar") == false {

		repoStr := GenerateRepo(table, st)

		forRepoStr, err := format.Source([]byte(repoStr))
		if err != nil {
			log.Errorf(err.Error())
		} else {
			err = ioutil.WriteFile(repoTarget, forRepoStr, 0666)
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
		log.Infof("create file  %s success", repoTarget)
	} else {
		log.Infof("not create file  %s", repoTarget)
	}

	inject := v.GetBool("inject")
	if inject == true {
		pwd, _ := os.Getwd()
		goPath := os.Getenv("GOPATH") + "/src/"
		//项目路径
		proPath := strings.Replace(pwd, goPath, "", -1)
		Inject("infra", st, pk,
			st+"Repo", proPath+"/internal/infra/repo")
	}

	return nil
}

func GenerateDao(tableName string, structName string, pkgName string,
	v *viper.Viper, genMysqlInfo generateMysqlInfo) ([]byte, error) {

	daoStr := `
package dao

import (
	"errors"
	"context"
	"{PROPATH}` + file_dir.GetParDir() + `/internal/domain/entity"
	"github.com/jinzhu/gorm"
	"github.com/jukylin/esim/mysql"
)


type ` + structName + `Dao struct{
	mysql mysql.MysqlClient
}

func New` + structName + `Dao() *` + structName + `Dao {
	dao := &` + structName + `Dao{
		mysql : mysql.NewMysqlClient(),
	}

	return dao
}


//主库
func (this *` + structName + `Dao) GetDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "` + pkgName + `").Table("` + tableName + `")
}

//从库
func (this *` + structName + `Dao) GetSlaveDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "` + pkgName + `_slave").Table("` + tableName + `")
}


//返回 自增id，错误
func (this *` + structName + `Dao) Create(ctx context.Context, ` + GetFirstToLower(structName) +
		` *entity.` + structName + `) (` + genMysqlInfo.priKeyType + `, error){
	db := this.GetDb(ctx).Create(` + GetFirstToLower(structName) + `)
	if db.Error != nil{
		return ` + genMysqlInfo.priKeyType + `(0), db.Error
	}else{
		return ` + genMysqlInfo.priKeyType + `(` + GetFirstToLower(structName) + `.ID), nil
	}
}

//ctx, "name = ?", "test"
func (this *` + structName + `Dao) Count(ctx context.Context, query interface{}, args ...interface{}) (int64, error){
	var count int64
	db := this.GetSlaveDb(ctx).Where(query, args...).Count(&count)
	if db.Error != nil{
		return count, db.Error
	}else{
		return count, nil
	}
}

// ctx, "id,name", "name = ?", "test"
// 可通过 HasData() 判断是否有数据
func (this *` + structName + `Dao) Find(ctx context.Context, squery , wquery interface{}, args ...interface{}) (entity.` + structName + `, error){
	var ` + GetFirstToLower(structName) + ` entity.` + structName + `
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&` + GetFirstToLower(structName) + `)
	if db.Error != nil{
		return ` + strings.ToLower(string(structName[0])) + `, db.Error
	}else{
		return ` + strings.ToLower(string(structName[0])) + `, nil
	}
}


// ctx, "id,name", "name = ?", "test"
//最多取10条
func (this *` + structName + `Dao) List(ctx context.Context, squery , wquery interface{}, args ...interface{}) ([]entity.` + structName + `, error){
	` + GetFirstToLower(structName) + `s := []entity.` + structName + `{}
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(10).Find(&` + GetFirstToLower(structName) + `s)
	if db.Error != nil{
		return ` + GetFirstToLower(structName) + `s, db.Error
	}else{
		return ` + GetFirstToLower(structName) + `s, nil
	}
}

func (this *` + structName + `Dao) DelById(ctx context.Context, id ` + genMysqlInfo.priKeyType + `) (bool, error){
	var del` + structName + ` entity.` + structName + `

	if del` + structName + `.DelKey() == ""{
		return false, errors.New("找不到 is_del / is_deleted / is_delete 字段")
	}

	del` + structName + `.ID = id
	db := this.GetDb(ctx).Update(map[string]interface{}{del` + structName + `.DelKey(): 1})
	if db.Error != nil{
		return false, db.Error
	}else{
		return true, nil
	}
}

//ctx, map[string]interface{}{"name": "hello"}, "name = ?", "test"
//返回影响数
func (this *` + structName + `Dao) Update(ctx context.Context, update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {
	db := this.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}

`

	return []byte(daoStr), nil
}

func GenerateRepo(tableName string, structName string) string {
	repoStr := `
package repo


import (
	"context"
	"{PROPATH}` + file_dir.GetParDir() + `/internal/domain/entity"
	"{PROPATH}` + file_dir.GetParDir() + `/internal/infra/repo/dao"
	"github.com/jukylin/esim/log"
)


type ` + structName + `Repo interface {
	FindById(context.Context, int64) entity.` + structName + `
}

type ` + tableName + `DbRepo struct{

	log log.Logger

	` + tableName + `Dao *dao.` + structName + `Dao
}

func New` + structName + `DbRepo(logger log.Logger) ` + structName + `DbRepo {
	repo := &` + tableName + `Repo{
		log : logger,
	}

	if repo.` + tableName + `Dao == nil{
		repo.` + tableName + `Dao = dao.New` + structName + `Dao()
	}


	return repo
}

func (this *` + tableName + `DbRepo) FindById(ctx context.Context, id int64) entity.` + structName + ` {
	var ` + tableName + ` entity.` + structName + `
	var err error

	` + tableName + `, err = this.` + tableName + `Dao.Find(ctx, "*", "id = ? ", id)

	if err != nil{
		this.log.Errorf(err.Error())
		return ` + tableName + `
	}

	return ` + tableName + `
}

`

	return repoStr
}

func GenInitFile(pkgName string, v *viper.Viper) string {
	initStr := `
package ` + pkgName + `

	import (
		"gopkg.in/go-playground/mold.v2/modifiers"
	)

	var (
	conform  = modifiers.New()
)
`

	return initStr
}

func camelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func GetFirstToLower(str string) string {
	return strings.ToLower(string(str[0]))
}

func Inject(structName string, fieldName, packageName, packageStrct string,
	importStr string) {

	infrDir := "./internal/infra/"

	infrFile := "infra.go"

	exists, err := file_dir.IsExistsFile(infrDir + infrFile)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if exists {
		src, err := ioutil.ReadFile(infrDir + infrFile)
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		formatSrc, err := format.Source([]byte(src))
		if err != nil {
			log.Errorf(err.Error())
			return
		}

		ExecGoFmt(infrFile, infrDir)

		//先整理下源文件
		ioutil.WriteFile(infrDir+infrFile, formatSrc, 0755)

		srcStr := string(formatSrc)

		source := handleInject(srcStr, "Infra",
			fieldName, packageName, packageStrct, importStr)

		//整理，写入
		//formatSrc, err = format.Source([]byte(source))
		//if err != nil{
		//	log.Errorf(err.Error())
		//	return
		//}

		//语法检查
		ioutil.WriteFile(infrDir+infrFile, []byte(source), 0755)

		ExecGoFmt(infrFile, infrDir)
		ExecWire(infrDir)

		log.Infof("注入成功")

	} else {
		log.Errorf("不存在 %s", infrDir+infrFile)
	}
}

func handleInject(srcStr string, structName string, fieldName, packageName, packageStrct string,
	importStr string) string {

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", srcStr, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var hasStruct bool
	var oldStruct string
	var newStruct string

	var oldImportStr string
	var newImportStr string

	provideFunc := getProvideFunc(packageName, packageStrct)

	var oldSet string
	var newSet string

	for _, decl := range f.Decls {

		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				oldImports := getOldImports(GenDecl)
				newImports := append(oldImports, "\""+importStr+"\"")
				oldImportStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
				newImportStr = getNewImportStr(newImports)
			}

			if GenDecl.Tok.String() == "type" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == structName {
							hasStruct = true
							oldStruct = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
							oldFields := GetOldFields(GenDecl, srcStr)
							newFields := append(oldFields, Field{Filed: packageStrct + " repo." + packageStrct})
							newStruct = getNewStruct(structName, newFields)
						}
					}
				}
			}

			if GenDecl.Tok.String() == "var" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.ValueSpec); ok {
						for _, name := range typeSpec.Names {
							if name.String() == "infraSet" {
								var oldArgs []string
								oldSet = srcStr[GenDecl.TokPos-1 : GenDecl.End()]
								oldArgs = append(oldArgs, "var infraSet = wire.NewSet(")
								oldArgs = append(oldArgs, getSet(GenDecl, srcStr)...)
								newArgs := append(oldArgs, "provide"+packageStrct+",")
								newArgs = append(newArgs, ")")
								newSet = getNewSet(newArgs)
							}
						}
					}
				}
			}
		}
	}

	//println(srcStr)
	if hasStruct == false {
		log.Errorf("不存在 %s", structName)
		return ""
	}

	srcStr = strings.Replace(srcStr, oldImportStr, newImportStr, -1)
	srcStr = strings.Replace(srcStr, oldStruct, newStruct, -1)
	srcStr = strings.Replace(srcStr, oldSet, newSet, -1)
	srcStr += provideFunc

	return srcStr
}

func getNewSet(args []string) string {
	var structStr string

	for _, a := range args {
		structStr += "	" + a + "\r\n"
	}
	return structStr
}

func getOldImports(GenDecl *ast.GenDecl) []string {
	var imports []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ImportSpec); ok {
			imports = append(imports, spec.Path.Value)
		}
	}

	return imports
}

type Field struct {
	Doc   []string
	Filed string
	Tag   string
}

func GetOldFields(GenDecl *ast.GenDecl, strSrc string) []Field {
	var fields []Field
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.TypeSpec); ok {
			if structType, ok := spec.Type.(*ast.StructType); ok {
				for _, astField := range structType.Fields.List {
					var field Field
					if astField.Doc != nil {
						for _, doc := range astField.Doc.List {
							field.Doc = append(field.Doc, doc.Text)
						}
					}

					if astField.Tag != nil {
						field.Tag = astField.Tag.Value
					}
					var name string
					if len(astField.Names) > 0 {
						name = astField.Names[0].String()
						field.Filed = name + " " + strSrc[astField.Type.Pos()-1:astField.Type.End()-1]
					} else {
						field.Filed = strSrc[astField.Type.Pos()-1 : astField.Type.End()-1]
					}

					fields = append(fields, field)
				}
			}
		}
	}

	return fields
}

func getSet(GenDecl *ast.GenDecl, srcStr string) []string {
	var args []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ValueSpec); ok {
			for _, value := range spec.Values {
				if callExpr, ok := value.(*ast.CallExpr); ok {
					for _, callArg := range callExpr.Args {
						if arg, ok := callArg.(*ast.CallExpr); ok {
							var x string
							var sel string
							if fun, ok := arg.Fun.(*ast.SelectorExpr); ok {
								x = fun.X.(*ast.Ident).String()
								sel = fun.Sel.String()
							}
							args = append(args, x+"."+sel+srcStr[arg.Lparen-1:arg.Rparen]+",")
						}

						if arg, ok := callArg.(*ast.Ident); ok {
							args = append(args, arg.String()+",")
						}
					}
				}
			}
		}
	}

	return args
}

func getNewImportStr(newImports []string) string {

	var importStr = "import (\r\n"

	for _, is := range newImports {
		importStr += "	" + is + "\r\n"
	}

	importStr += ")\r\n"

	return importStr
}

func getProvideFunc(packageName, packageStrct string) string {
	funcStr := `
func provide` + packageStrct + `(esim *container.Esim) repo.` + packageStrct + ` {
	return repo.New` + packageStrct + `(esim.Log)
}`

	return funcStr
}

func getNewStruct(name string, fields []Field) string {
	var structStr string
	structStr = " type " + name + " struct {\r\n"

	for _, f := range fields {
		if len(f.Doc) > 0 {
			for _, d := range f.Doc {
				structStr += "	" + d + "\r\n"
			}
		}
		structStr += "	" + f.Filed + "\r\n"
		structStr += "\r\n"
	}

	structStr += "}\r\n"

	return structStr
}

func GetFirstToUpper(str string) string {
	return strings.ToUpper(string(str[0])) + str[1:]
}

func ExecGoFmt(file string, dir string) {
	cmd_line := fmt.Sprintf("go fmt %s", file)

	log.Infof(cmd_line)

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf(err.Error())
	}
}

func ExecWire(dir string) {
	cmd_line := fmt.Sprintf("wire")

	log.Infof("dir %s, %s", dir, cmd_line)

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf(err.Error())
	}
}