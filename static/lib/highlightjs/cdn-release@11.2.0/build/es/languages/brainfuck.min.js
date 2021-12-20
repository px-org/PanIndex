var hljsGrammar=(()=>{"use strict";return e=>{const a={className:"literal",
begin:/[+-]/,relevance:0};return{name:"Brainfuck",aliases:["bf"],
contains:[e.COMMENT("[^\\[\\]\\.,\\+\\-<> \r\n]","[\\[\\]\\.,\\+\\-<> \r\n]",{
returnEnd:!0,relevance:0}),{className:"title",begin:"[\\[\\]]",relevance:0},{
className:"string",begin:"[\\.,]",relevance:0},{begin:/(?:\+\+|--)/,contains:[a]
},a]}}})();export default hljsGrammar;