const Koa = require('koa');
const app = new Koa();

app.use(async ctx => {
  ctx.body = "hello world 4"
})

app.listen(10080, () => {
  console.log("listen at 10080")
})
