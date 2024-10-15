# gopherlox
![image](https://github.com/user-attachments/assets/f3576184-a90a-4bca-a1d8-4632e27fcc8f)
A go implementation of the lox programming language

## TODOs
### Lexer
- [x] implement all token types
- [x] implement all keywords
- [x] implement all operators
- [x] implement all literals
- [x] implement identifiers
- [x] implement numbers (floating point)
- [x] implement strings
- [x] implement comments
- [x] skip whitespaces
- [ ] implement multi line comments using /* */
- [ ] fix multiline string parsing

### Parser
- [ ] implement the following grammar
```
expression → literal
| unary
| binary
| grouping ;
literal → NUMBER | STRING | "true" | "false" | "nil" ;
grouping → "(" expression ")" ;
unary → ( "-" | "!" ) expression ;
binary → expression operator expression ;
operator → "==" | "!=" | "<" | "<=" | 
```