(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["default-account"],{2909:function(t,e,n){"use strict";n.d(e,"a",(function(){return u}));var i=n("6b75");function o(t){if(Array.isArray(t))return Object(i["a"])(t)}n("a4d3"),n("e01a"),n("d3b7"),n("d28b"),n("3ca3"),n("ddb0"),n("a630");function r(t){if("undefined"!==typeof Symbol&&null!=t[Symbol.iterator]||null!=t["@@iterator"])return Array.from(t)}var a=n("06c5");function s(){throw new TypeError("Invalid attempt to spread non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}function u(t){return o(t)||r(t)||Object(a["a"])(t)||s()}},"38ef":function(t,e,n){"use strict";n.r(e);var i=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("v-menu",{attrs:{bottom:"",left:"","min-width":"200","offset-y":"",origin:"top right",transition:"scale-transition"},scopedSlots:t._u([{key:"activator",fn:function(e){var i=e.attrs,o=e.on;return[n("v-btn",t._g(t._b({staticClass:"ml-2",attrs:{"min-width":"0",text:""}},"v-btn",i,!1),o),[n("v-icon",[t._v("mdi-account")])],1)]}}])},[n("v-list",{attrs:{tile:!1,flat:"",nav:""}},[t._l(t.profile,(function(e,i){return[e.divider?n("v-divider",{key:"divider-"+i,staticClass:"mb-2 mt-2"}):n("v-list-item",{key:"item-"+i,attrs:{to:e.link}},[n("v-list-item-title",{domProps:{textContent:t._s(e.title)}})],1)]}))],2)],1)},o=[],r={name:"DefaultAccount",data:function(){return{profile:[{title:"Profile",link:"/pages/profile"},{divider:!0},{title:"Log out",link:"/auth/logout"}]}}},a=r,s=n("2877"),u=n("6544"),c=n.n(u),d=n("8336"),l=n("ce7e"),f=n("132d"),v=n("8860"),m=n("da13"),b=n("5d23"),p=n("e449"),w=Object(s["a"])(a,i,o,!1,null,null,null);e["default"]=w.exports;c()(w,{VBtn:d["a"],VDivider:l["a"],VIcon:f["a"],VList:v["a"],VListItem:m["a"],VListItemTitle:b["b"],VMenu:p["a"]})},dc22:function(t,e,n){"use strict";function i(t,e){var n=e.value,i=e.options||{passive:!0};window.addEventListener("resize",n,i),t._onResize={callback:n,options:i},e.modifiers&&e.modifiers.quiet||n()}function o(t){if(t._onResize){var e=t._onResize,n=e.callback,i=e.options;window.removeEventListener("resize",n,i),delete t._onResize}}var r={inserted:i,unbind:o};e["a"]=r},dd89:function(t,e,n){"use strict";function i(t){if("function"!==typeof t.getRootNode){while(t.parentNode)t=t.parentNode;return t!==document?null:document}var e=t.getRootNode();return e!==document&&e.getRootNode({composed:!0})!==document?null:e}n.d(e,"a",(function(){return i}))}}]);
//# sourceMappingURL=default-account.83bbad60.js.map