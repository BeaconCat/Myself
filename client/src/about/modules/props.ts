import type { AboutModule } from '../../stores/config';
import type { Span } from '../types';

/** 所有模块渲染器的统一 props（由 AboutModules 注入已迁移的数据与生效的 span / variant / 标题） */
export interface ModProps {
  mod: AboutModule;
  variant: string;
  span: Span;
  title: string;
  /** chapter 自动编号（01、02…） */
  no?: string;
}
