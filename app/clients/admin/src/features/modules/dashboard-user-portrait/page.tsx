import { Empty } from "antd";
import { m } from "#/paraglide/messages";

// 用户画像本轮只挂菜单占位：画像口径依赖上游业务域（测评 / 支付 / 用户标签），
// 这些域尚未在本仓落地，故给出明确空态而不是空表。
export function UserPortraitPage() {
  return (
    <div className="flex h-full items-center justify-center p-4">
      <Empty
        description={m.dashboard_overview_user_portrait_empty()}
        image={Empty.PRESENTED_IMAGE_SIMPLE}
      />
    </div>
  );
}
