var hljsGrammar=(()=>{"use strict";return e=>{const n={
keyword:["rec","with","let","in","inherit","assert","if","else","then"],
literal:["true","false","or","and","null"],
built_in:["import","abort","baseNameOf","dirOf","isNull","builtins","map","removeAttrs","throw","toString","derivation"]
},r={className:"subst",begin:/\$\{/,end:/\}/,keywords:n},t={className:"string",
contains:[r],variants:[{begin:"''",end:"''"},{begin:'"',end:'"'}]
},a=[e.NUMBER_MODE,e.HASH_COMMENT_MODE,e.C_BLOCK_COMMENT_MODE,t,{
begin:/[a-zA-Z0-9-_]+(\s*=)/,returnBegin:!0,relevance:0,contains:[{
className:"attr",begin:/\S+/}]}];return r.contains=a,{name:"Nix",
aliases:["nixos"],keywords:n,contains:a}}})();export default hljsGrammar;