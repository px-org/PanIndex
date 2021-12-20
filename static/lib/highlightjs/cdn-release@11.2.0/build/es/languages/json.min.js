var hljsGrammar=(()=>{"use strict";return a=>({name:"JSON",contains:[{
className:"attr",begin:/"(\\.|[^\\"\r\n])*"(?=\s*:)/,relevance:1.01},{
match:/[{}[\],:]/,className:"punctuation",relevance:0},a.QUOTE_STRING_MODE,{
beginKeywords:"true false null"
},a.C_NUMBER_MODE,a.C_LINE_COMMENT_MODE,a.C_BLOCK_COMMENT_MODE],illegal:"\\S"})
})();export default hljsGrammar;