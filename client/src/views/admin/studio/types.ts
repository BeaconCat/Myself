/** Studio 共享类型 */
export interface MenuItem {
  icon: string;
  label: string;
  danger?: boolean;
  /** 在该项之前画一条分隔线 */
  divider?: boolean;
  run: () => void;
}

/** 素材被引用的位置 */
export interface MediaRef {
  kind: 'post' | 'note';
  id: number;
  title: string;
}
