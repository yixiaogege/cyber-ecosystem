// 数据可视化配色：取值对齐 manage 原型 dashboard.html，且刻意不做主题化——
// 序列色一旦跟随皮肤变化，「新增注册」与「完成测评」的对照关系就不再稳定。
// 语义固定：s1 = 主序列（新增注册 / 被邀请用户 / 邀请人条形），s2 = 对照序列（完成测评 / 自然流量）。
export const S1 = "#2563eb";
export const S2 = "#b45309";

// 有序阶段（漏斗 5 阶、裂变层级 4 档）→ 序数色阶。阶段本身有序，故用同色相深浅而非分类色。
export const ORDINAL_FUNNEL = ["#1e40af", "#2563eb", "#3b82f6", "#60a5fa", "#93c5fd"];
export const ORDINAL_LEVEL = ["#1e40af", "#2563eb", "#3b82f6", "#93c5fd"];

// 图表骨架（网格线、坐标轴文字）——中性灰，不参与语义。
export const GRID = "#e5e7eb";
export const AXIS_TEXT = "#8c8c8c";
export const TRACK = "var(--ant-color-fill-tertiary, #f0f0f0)";
