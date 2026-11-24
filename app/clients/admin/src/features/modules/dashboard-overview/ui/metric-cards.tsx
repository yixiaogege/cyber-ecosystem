import type { MetricCard } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { Link } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";
import { fixed, fmt } from "./format";

// toJson 把 proto 枚举转成字符串名（如 "CARD_DRILL_USER_LIST"），
// 所以 Record 键必须用字符串而非 TS 枚举数字值。
// enumKey 把 TS 枚举值（number）在运行时已被 toJson 替换成的字符串安全取出。
function enumKey(e: number): string {
  return e as unknown as string;
}

// 下钻目标：本仓只存在用户列表；测试记录 / 订单属于上游业务域（本轮不建），
// 因此渲染为不可点的占位卡而非死链——原型的四个入口保持可见。
const DRILL_TARGET: Record<string, string | undefined> = {
  CARD_DRILL_USER_LIST: "/dashboard/system/users",
};

const DRILL_LABEL: Record<string, () => string> = {
  CARD_DRILL_UNSPECIFIED: () => m.dashboard_overview_drill_user_list(),
  CARD_DRILL_USER_LIST: () => m.dashboard_overview_drill_user_list(),
  CARD_DRILL_TEST_RECORD: () => m.dashboard_overview_drill_test_record(),
  CARD_DRILL_ORDER: () => m.dashboard_overview_drill_order(),
};

const CARD_LABEL: Record<string, () => string> = {
  METRIC_KEY_UNSPECIFIED: () => m.dashboard_overview_card_register(),
  METRIC_KEY_REGISTER: () => m.dashboard_overview_card_register(),
  METRIC_KEY_TEST: () => m.dashboard_overview_card_test(),
  METRIC_KEY_ACTIVE: () => m.dashboard_overview_card_active(),
  METRIC_KEY_PAY: () => m.dashboard_overview_card_pay(),
};

// 副标签按口径分派；缺数值时整行不渲染（返回 undefined）。
const SUB_TEXT: Record<string, (card: MetricCard, isDay: boolean) => string | undefined> = {
  SUB_METRIC_KIND_UNSPECIFIED: () => undefined,
  SUB_METRIC_KIND_DELTA_PREV: (card, isDay) => {
    if (card.subValue === undefined) return undefined;
    const value = `${card.subValue > 0 ? "+" : ""}${fixed(card.subValue, 0)}%`;
    return isDay
      ? m.dashboard_overview_card_sub_delta_day({ value })
      : m.dashboard_overview_card_sub_delta({ value });
  },
  SUB_METRIC_KIND_AVG_PER_USER: (card) =>
    card.subValue === undefined
      ? undefined
      : m.dashboard_overview_card_sub_avg({ value: fixed(card.subValue) }),
  SUB_METRIC_KIND_PEAK: (card) =>
    card.subValue === undefined
      ? undefined
      : m.dashboard_overview_card_sub_peak({
          days: card.subNote ?? "1",
          value: fmt(card.subValue),
        }),
  SUB_METRIC_KIND_MANUAL: (card) =>
    card.subValue === undefined
      ? undefined
      : m.dashboard_overview_card_sub_manual({ value: fmt(card.subValue) }),
};

// subText 把副标签的三要素（数值 / 口径 / 窗口）拼成一行。
function subText(card: MetricCard, isDay: boolean): string | undefined {
  const fn = SUB_TEXT[enumKey(card.subKind)];
  return fn ? fn(card, isDay) : undefined;
}

function drillLabel(drill: number): string {
  const fn = DRILL_LABEL[enumKey(drill)] ?? DRILL_LABEL.CARD_DRILL_UNSPECIFIED;
  return fn ? fn() : "";
}

function cardLabel(key: number): string {
  const fn = CARD_LABEL[enumKey(key)] ?? CARD_LABEL.METRIC_KEY_UNSPECIFIED;
  return fn ? fn() : "";
}

export function MetricCards({
  cards,
  isDay,
}: Readonly<{ cards: readonly MetricCard[]; isDay: boolean }>) {
  return (
    <div className="grid grid-cols-4 gap-3">
      {cards.map((card) => {
        const target = DRILL_TARGET[enumKey(card.drill)];
        const hint = drillLabel(card.drill);
        const sub = subText(card, isDay);
        const body = (
          <>
            <div className="text-ink-secondary text-xs">{cardLabel(card.key)}</div>
            <div className="mt-1 font-black text-[26px] text-primary tabular-nums leading-tight">
              {fmt(Number(card.value))}
            </div>
            <div className="mt-1 text-ink-tertiary text-xs">
              {sub ? `${sub} · ` : ""}
              {target
                ? m.dashboard_overview_card_drill({ target: hint })
                : m.dashboard_overview_card_drill_off({ target: hint })}
            </div>
          </>
        );
        const shell = "block rounded-xl border border-line bg-container p-3.5 transition-colors";
        if (!target) {
          return (
            <div aria-disabled className={`${shell} opacity-60`} key={card.key}>
              {body}
            </div>
          );
        }
        return (
          <Link className={`${shell} hover:border-line-hover`} key={card.key} to={target}>
            {body}
          </Link>
        );
      })}
    </div>
  );
}
