/** Studio 共享类型 */
export interface MenuItem {
  icon: string;
  label: string;
  danger?: boolean;
  /** 在该项之前画一条分隔线 */
  divider?: boolean;
  run: () => void;
}

/** 素材被引用的位置：文章 / 随想 / 身份（站点 logo、头像、形象图、名片头图）/ 关于页模块 */
export interface MediaRef {
  kind: 'post' | 'note' | 'identity' | 'about';
  id: number;
  title: string;
}
