// Package ast defines the Abstract Syntax Tree for NilLang.
package ast

import "github.com/joysriramsarkar/alap-framework/compiler/lexer"

// ─── Node interfaces ──────────────────────────────────────────────────────────

// Node is the base type for all AST nodes.
type Node interface {
	NodePos() lexer.Position
}

// Statement is a Node that does not produce a value.
type Statement interface {
	Node
	stmtNode()
}

// Expression is a Node that produces a value.
type Expression interface {
	Node
	exprNode()
}

// Declaration is a top-level statement.
type Declaration interface {
	Statement
	declNode()
}

// ─── Program ─────────────────────────────────────────────────────────────────

type Program struct {
	Statements []Statement
	Pos        lexer.Position
}

func (p *Program) NodePos() lexer.Position { return p.Pos }

// ─── Identifiers ─────────────────────────────────────────────────────────────

type Identifier struct {
	Name string
	Pos  lexer.Position
}

func (i *Identifier) NodePos() lexer.Position { return i.Pos }
func (i *Identifier) exprNode()               {}

// ─── Type annotations ────────────────────────────────────────────────────────

type TypeAnnotation struct {
	Name     string
	Nullable bool
	Array    bool
	Generic  []*TypeAnnotation
	Union    []*TypeAnnotation
	Tuple    []*TypeAnnotation
	Func     *FuncType
	Pos      lexer.Position
}

func (t *TypeAnnotation) NodePos() lexer.Position { return t.Pos }

type FuncType struct {
	Params []*TypeAnnotation
	Return *TypeAnnotation
	Pos    lexer.Position
}

// ─── Literals ────────────────────────────────────────────────────────────────

type IntLiteral struct {
	Value string
	Pos   lexer.Position
}

func (n *IntLiteral) NodePos() lexer.Position { return n.Pos }
func (n *IntLiteral) exprNode()               {}

type FloatLiteral struct {
	Value string
	Pos   lexer.Position
}

func (n *FloatLiteral) NodePos() lexer.Position { return n.Pos }
func (n *FloatLiteral) exprNode()               {}

type StringLiteral struct {
	Value string
	Pos   lexer.Position
}

func (n *StringLiteral) NodePos() lexer.Position { return n.Pos }
func (n *StringLiteral) exprNode()               {}

type BoolLiteral struct {
	Value bool
	Pos   lexer.Position
}

func (n *BoolLiteral) NodePos() lexer.Position { return n.Pos }
func (n *BoolLiteral) exprNode()               {}

type NullLiteral struct{ Pos lexer.Position }

func (n *NullLiteral) NodePos() lexer.Position { return n.Pos }
func (n *NullLiteral) exprNode()               {}

type ArrayLiteral struct {
	Elements []Expression
	Pos      lexer.Position
}

func (n *ArrayLiteral) NodePos() lexer.Position { return n.Pos }
func (n *ArrayLiteral) exprNode()               {}

type ObjectLiteral struct {
	Fields []ObjectField
	Pos    lexer.Position
}

type ObjectField struct {
	Name  string
	Value Expression
	Pos   lexer.Position
}

func (n *ObjectLiteral) NodePos() lexer.Position { return n.Pos }
func (n *ObjectLiteral) exprNode()               {}

// ─── Expressions ─────────────────────────────────────────────────────────────

type BinaryExpr struct {
	Op    string
	Left  Expression
	Right Expression
	Pos   lexer.Position
}

func (n *BinaryExpr) NodePos() lexer.Position { return n.Pos }
func (n *BinaryExpr) exprNode()               {}

type UnaryExpr struct {
	Op      string
	Operand Expression
	Prefix  bool
	Pos     lexer.Position
}

func (n *UnaryExpr) NodePos() lexer.Position { return n.Pos }
func (n *UnaryExpr) exprNode()               {}

type AssignExpr struct {
	Op    string
	Left  Expression
	Right Expression
	Pos   lexer.Position
}

func (n *AssignExpr) NodePos() lexer.Position { return n.Pos }
func (n *AssignExpr) exprNode()               {}

type CallExpr struct {
	Callee    Expression
	TypeArgs  []*TypeAnnotation
	Arguments []Expression
	Pos       lexer.Position
}

func (n *CallExpr) NodePos() lexer.Position { return n.Pos }
func (n *CallExpr) exprNode()               {}

type IndexExpr struct {
	Object Expression
	Index  Expression
	Pos    lexer.Position
}

func (n *IndexExpr) NodePos() lexer.Position { return n.Pos }
func (n *IndexExpr) exprNode()               {}

type MemberExpr struct {
	Object   Expression
	Property string
	Optional bool
	Pos      lexer.Position
}

func (n *MemberExpr) NodePos() lexer.Position { return n.Pos }
func (n *MemberExpr) exprNode()               {}

type ArrowFuncExpr struct {
	Params     []*Param
	ReturnType *TypeAnnotation
	Body       Node
	Async      bool
	Pos        lexer.Position
}

func (n *ArrowFuncExpr) NodePos() lexer.Position { return n.Pos }
func (n *ArrowFuncExpr) exprNode()               {}

type NewExpr struct {
	Constructor Expression
	TypeArgs    []*TypeAnnotation
	Arguments   []Expression
	Pos         lexer.Position
}

func (n *NewExpr) NodePos() lexer.Position { return n.Pos }
func (n *NewExpr) exprNode()               {}

type AwaitExpr struct {
	Operand Expression
	Pos     lexer.Position
}

func (n *AwaitExpr) NodePos() lexer.Position { return n.Pos }
func (n *AwaitExpr) exprNode()               {}

type TryExpr struct {
	Operand Expression
	Pos     lexer.Position
}

func (n *TryExpr) NodePos() lexer.Position { return n.Pos }
func (n *TryExpr) exprNode()               {}

type ChanReceiveExpr struct {
	Channel Expression
	Pos     lexer.Position
}

func (n *ChanReceiveExpr) NodePos() lexer.Position { return n.Pos }
func (n *ChanReceiveExpr) exprNode()               {}

type TypeAssertExpr struct {
	Value Expression
	Type  *TypeAnnotation
	Pos   lexer.Position
}

func (n *TypeAssertExpr) NodePos() lexer.Position { return n.Pos }
func (n *TypeAssertExpr) exprNode()               {}

type TernaryExpr struct {
	Condition   Expression
	Consequent  Expression
	Alternative Expression
	Pos         lexer.Position
}

func (n *TernaryExpr) NodePos() lexer.Position { return n.Pos }
func (n *TernaryExpr) exprNode()               {}

// ─── Statements ──────────────────────────────────────────────────────────────

type BlockStatement struct {
	Body []Statement
	Pos  lexer.Position
}

func (n *BlockStatement) NodePos() lexer.Position { return n.Pos }
func (n *BlockStatement) stmtNode()               {}

type ExprStatement struct {
	Expr Expression
	Pos  lexer.Position
}

func (n *ExprStatement) NodePos() lexer.Position { return n.Pos }
func (n *ExprStatement) stmtNode()               {}

type VarDecl struct {
	Kind string
	Name string
	Type *TypeAnnotation
	Init Expression
	Pos  lexer.Position
}

func (n *VarDecl) NodePos() lexer.Position { return n.Pos }
func (n *VarDecl) stmtNode()               {}
func (n *VarDecl) declNode()               {}

type ReturnStatement struct {
	Value Expression
	Pos   lexer.Position
}

func (n *ReturnStatement) NodePos() lexer.Position { return n.Pos }
func (n *ReturnStatement) stmtNode()               {}

type IfStatement struct {
	Condition   Expression
	Consequent  *BlockStatement
	Alternative Statement
	Pos         lexer.Position
}

func (n *IfStatement) NodePos() lexer.Position { return n.Pos }
func (n *IfStatement) stmtNode()               {}

type WhileStatement struct {
	Condition Expression
	Body      *BlockStatement
	Pos       lexer.Position
}

func (n *WhileStatement) NodePos() lexer.Position { return n.Pos }
func (n *WhileStatement) stmtNode()               {}

type ForStatement struct {
	Init      Statement
	Condition Expression
	Update    Expression
	Body      *BlockStatement
	Pos       lexer.Position
}

func (n *ForStatement) NodePos() lexer.Position { return n.Pos }
func (n *ForStatement) stmtNode()               {}

type ForInStatement struct {
	VarKind  string
	VarName  string
	VarType  *TypeAnnotation
	Iterable Expression
	Body     *BlockStatement
	Pos      lexer.Position
}

func (n *ForInStatement) NodePos() lexer.Position { return n.Pos }
func (n *ForInStatement) stmtNode()               {}

type LoopStatement struct {
	Body *BlockStatement
	Pos  lexer.Position
}

func (n *LoopStatement) NodePos() lexer.Position { return n.Pos }
func (n *LoopStatement) stmtNode()               {}

type BreakStatement struct{ Pos lexer.Position }

func (n *BreakStatement) NodePos() lexer.Position { return n.Pos }
func (n *BreakStatement) stmtNode()               {}

type ContinueStatement struct{ Pos lexer.Position }

func (n *ContinueStatement) NodePos() lexer.Position { return n.Pos }
func (n *ContinueStatement) stmtNode()               {}

type ThrowStatement struct {
	Value Expression
	Pos   lexer.Position
}

func (n *ThrowStatement) NodePos() lexer.Position { return n.Pos }
func (n *ThrowStatement) stmtNode()               {}

type TryCatchStatement struct {
	Try     *BlockStatement
	Catch   *CatchClause
	Finally *BlockStatement
	Pos     lexer.Position
}

type CatchClause struct {
	Param *Param
	Body  *BlockStatement
}

func (n *TryCatchStatement) NodePos() lexer.Position { return n.Pos }
func (n *TryCatchStatement) stmtNode()               {}

type MatchStatement struct {
	Subject Expression
	Arms    []*MatchArm
	Pos     lexer.Position
}

type MatchArm struct {
	Pattern Expression
	Guard   Expression
	Body    *BlockStatement
}

func (n *MatchStatement) NodePos() lexer.Position { return n.Pos }
func (n *MatchStatement) stmtNode()               {}

type TaskStatement struct {
	Body *BlockStatement
	Pos  lexer.Position
}

func (n *TaskStatement) NodePos() lexer.Position { return n.Pos }
func (n *TaskStatement) stmtNode()               {}

type SpawnStatement struct {
	Call Expression
	Pos  lexer.Position
}

func (n *SpawnStatement) NodePos() lexer.Position { return n.Pos }
func (n *SpawnStatement) stmtNode()               {}

// ChanSendStatement: ch <- val
// It is both a Statement and an Expression so it can appear inline.
type ChanSendStatement struct {
	Channel Expression
	Value   Expression
	Pos     lexer.Position
}

func (n *ChanSendStatement) NodePos() lexer.Position { return n.Pos }
func (n *ChanSendStatement) stmtNode()               {}
func (n *ChanSendStatement) exprNode()               {}

// ─── Declarations ────────────────────────────────────────────────────────────

type Param struct {
	Name         string
	Type         *TypeAnnotation
	Optional     bool
	DefaultValue Expression
	Rest         bool
	Pos          lexer.Position
}

type FunctionDecl struct {
	Name       string
	TypeParams []string
	Params     []*Param
	ReturnType *TypeAnnotation
	Body       *BlockStatement
	Async      bool
	Decorators []*Decorator
	Pos        lexer.Position
}

func (n *FunctionDecl) NodePos() lexer.Position { return n.Pos }
func (n *FunctionDecl) stmtNode()               {}
func (n *FunctionDecl) declNode()               {}

type ClassDecl struct {
	Name       string
	TypeParams []string
	SuperClass string
	Implements []string
	Members    []ClassMember
	Decorators []*Decorator
	Pos        lexer.Position
}

func (n *ClassDecl) NodePos() lexer.Position { return n.Pos }
func (n *ClassDecl) stmtNode()               {}
func (n *ClassDecl) declNode()               {}

type ClassMember interface {
	memberNode()
}

type FieldMember struct {
	Name       string
	Type       *TypeAnnotation
	Init       Expression
	Access     string
	Static     bool
	Readonly   bool
	Optional   bool
	Decorators []*Decorator
	Pos        lexer.Position
}

func (n *FieldMember) memberNode() {}

type MethodMember struct {
	Name       string
	Params     []*Param
	ReturnType *TypeAnnotation
	Body       *BlockStatement
	Access     string
	Static     bool
	Async      bool
	Abstract   bool
	Decorators []*Decorator
	Pos        lexer.Position
}

func (n *MethodMember) memberNode() {}

type ConstructorMember struct {
	Params []*Param
	Body   *BlockStatement
	Pos    lexer.Position
}

func (n *ConstructorMember) memberNode() {}

type StructDecl struct {
	Name       string
	TypeParams []string
	Fields     []*StructField
	Decorators []*Decorator
	Pos        lexer.Position
}

type StructField struct {
	Name     string
	Type     *TypeAnnotation
	Optional bool
	Pos      lexer.Position
}

func (n *StructDecl) NodePos() lexer.Position { return n.Pos }
func (n *StructDecl) stmtNode()               {}
func (n *StructDecl) declNode()               {}

type InterfaceDecl struct {
	Name       string
	TypeParams []string
	Methods    []*InterfaceMethod
	Fields     []*StructField
	Pos        lexer.Position
}

type InterfaceMethod struct {
	Name       string
	Params     []*Param
	ReturnType *TypeAnnotation
	Async      bool
	Pos        lexer.Position
}

func (n *InterfaceDecl) NodePos() lexer.Position { return n.Pos }
func (n *InterfaceDecl) stmtNode()               {}
func (n *InterfaceDecl) declNode()               {}

type EnumDecl struct {
	Name    string
	Members []*EnumMember
	Pos     lexer.Position
}

type EnumMember struct {
	Name  string
	Value Expression
}

func (n *EnumDecl) NodePos() lexer.Position { return n.Pos }
func (n *EnumDecl) stmtNode()               {}
func (n *EnumDecl) declNode()               {}

type TypeDecl struct {
	Name       string
	TypeParams []string
	Type       *TypeAnnotation
	Pos        lexer.Position
}

func (n *TypeDecl) NodePos() lexer.Position { return n.Pos }
func (n *TypeDecl) stmtNode()               {}
func (n *TypeDecl) declNode()               {}

type ImportDecl struct {
	Names []ImportName
	From  string
	Pos   lexer.Position
}

type ImportName struct {
	Name  string
	Alias string
}

func (n *ImportDecl) NodePos() lexer.Position { return n.Pos }
func (n *ImportDecl) stmtNode()               {}
func (n *ImportDecl) declNode()               {}

type ExportDecl struct {
	Declaration Statement
	Names       []ImportName
	From        string
	Default     bool
	Pos         lexer.Position
}

func (n *ExportDecl) NodePos() lexer.Position { return n.Pos }
func (n *ExportDecl) stmtNode()               {}
func (n *ExportDecl) declNode()               {}

// ─── Decorators ──────────────────────────────────────────────────────────────

type Decorator struct {
	Name string
	Args []Expression
	Pos  lexer.Position
}

// ─── UI nodes ────────────────────────────────────────────────────────────────

type ComponentDecl struct {
	Name        string
	StateFields []*UIStateField
	PropFields  []*UIPropField
	BuildBody   *UINode
	Methods     []*MethodMember
	Lifecycle   []*MethodMember
	IsEntry     bool
	Pos         lexer.Position
}

func (n *ComponentDecl) NodePos() lexer.Position { return n.Pos }
func (n *ComponentDecl) stmtNode()               {}
func (n *ComponentDecl) declNode()               {}

type UIStateField struct {
	Name string
	Type *TypeAnnotation
	Init Expression
	Pos  lexer.Position
}

type UIPropField struct {
	Name string
	Type *TypeAnnotation
	Init Expression
	Pos  lexer.Position
}

type UINode struct {
	Widget        string
	Args          []Expression
	Modifiers     []*UIModifier
	Children      []*UINode
	EventHandlers []*UIEvent
	Pos           lexer.Position
}

type UIModifier struct {
	Name string
	Args []Expression
}

type UIEvent struct {
	Name    string
	Handler Expression
}

type ActorDecl struct {
	Name     string
	Fields   []*StructField
	Handlers []*ActorHandler
	Pos      lexer.Position
}

type ActorHandler struct {
	Name   string
	Params []*Param
	Body   *BlockStatement
	Pos    lexer.Position
}

func (n *ActorDecl) NodePos() lexer.Position { return n.Pos }
func (n *ActorDecl) stmtNode()               {}
func (n *ActorDecl) declNode()               {}

type StoreDecl struct {
	Name    string
	Fields  []*StructField
	Methods []*MethodMember
	Pos     lexer.Position
}

func (n *StoreDecl) NodePos() lexer.Position { return n.Pos }
func (n *StoreDecl) stmtNode()               {}
func (n *StoreDecl) declNode()               {}
