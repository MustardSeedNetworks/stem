/**
 * SidebarChrome — the rail's header: the product mark and name.
 *
 * Split out of Sidebar.tsx when UI-STEM-9 moved the shell's chrome into the
 * rail. The footer is the shared RailFooter, fed props from Sidebar.tsx; the
 * connection status and the collapse control live there, not here.
 */
import { useTranslation } from 'react-i18next';
import { ProductMark } from './ProductMark';

export interface SidebarHeaderProps {
  collapsed: boolean;
}

export function SidebarHeader({ collapsed }: SidebarHeaderProps) {
  const { t } = useTranslation();
  return (
    <div
      className={`flex items-center ${
        collapsed ? 'justify-center' : 'justify-start'
      } px-3 py-4 border-b border-hairline`}
    >
      <div className="flex items-center gap-compact">
        <ProductMark />
        {!collapsed ? (
          <span className="font-display font-bold text-lg text-text-primary tracking-tight">
            {t('app.title')}
          </span>
        ) : null}
      </div>
    </div>
  );
}
