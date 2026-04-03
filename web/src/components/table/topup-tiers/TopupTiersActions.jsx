import { Button } from '@douyinfe/semi-ui';
import { IconPlus } from '@douyinfe/semi-icons';

export default function TopupTiersActions({ openCreate, t }) {
  return (
    <div style={{ display: 'flex', gap: 8 }}>
      <Button icon={<IconPlus />} theme='solid' onClick={openCreate}>
        {t('创建档位')}
      </Button>
    </div>
  );
}
