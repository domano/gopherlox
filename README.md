# gopherlox
![image](https://github.com/user-attachments/assets/f3576184-a90a-4bca-a1d8-4632e27fcc8f)
A goation of the lox programming language

## TODOs
### Lexer
- [x] all token types
- [x] all keywords
- [x] all operators
- [x] all literals
- [x] identifiers
- [x] numbers (floating point)
- [x] strings
- [x] comments
- [x] skip whitespaces
- [ ] multi line comments using /* */
- [ ] fix multiline string parsing
- [ ] property based testing

### Parser
- [ ] basic grammar
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