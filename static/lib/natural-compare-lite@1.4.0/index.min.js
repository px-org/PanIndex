/*! litejs.com/MIT-LICENSE.txt */
var naturalCompare=function(g,h){function f(b,c,a){if(a){for(d=c;a=f(b,d),76>a&&65<a;)++d;return+b.slice(c-1,d)}a=m&&m.indexOf(b.charAt(c));return-1<a?a+76:(a=b.charCodeAt(c)||0,45>a||127<a)?a:46>a?65:48>a?a-1:58>a?a+18:65>a?a-11:91>a?a+11:97>a?a-37:123>a?a+5:a-63}var d,c,b=1,k=0,l=0,m=String.alphabet;if((g+="")!=(h+=""))for(;b;)if(c=f(g,k++),b=f(h,l++),76>c&&76>b&&66<c&&66<b&&(c=f(g,k,k),b=f(h,l,k=d),l=d),c!=b)return c<b?-1:1;return 0};
try{module.exports=naturalCompare}catch(e){String.naturalCompare=naturalCompare};
