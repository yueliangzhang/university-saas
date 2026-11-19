// V17.5 滚动诊断：复制整段到浏览器 DevTools Console 运行（只读）
(() => {
  const result = [];
  const selectors = ['html','body','#app','.student-portal','.teacher-portal','.portal-layout','.portal-main','.portal-content','.main','.content','.el-main'];
  for (const selector of selectors) {
    document.querySelectorAll(selector).forEach(el => {
      const s = getComputedStyle(el);
      result.push({selector, className: String(el.className).slice(0,80), position:s.position,
        height:s.height, minHeight:s.minHeight, overflowY:s.overflowY,
        clientHeight:el.clientHeight, scrollHeight:el.scrollHeight,
        scrollTop:el.scrollTop, canScroll:el.scrollHeight>el.clientHeight});
    });
  }
  console.table(result);
  console.log('viewport', {innerHeight:window.innerHeight, scrollY:window.scrollY,
    documentScrollHeight:document.documentElement.scrollHeight});
})();
