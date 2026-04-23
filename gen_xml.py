import uuid

def gen_id():
    return str(uuid.uuid4()).replace('-', '')[:20]

def create_table(name, fields, x, y, width=180):
    table_id = gen_id()
    height = 30 + len(fields) * 30
    xml = f'''        <mxCell id="{table_id}" value="{name}" style="shape=table;startSize=30;container=1;collapsible=1;childLayout=tableLayout;fixedRows=1;rowLines=0;fontStyle=1;align=center;resizeLast=1;html=1;" vertex="1" parent="1">
          <mxGeometry x="{x}" y="{y}" width="{width}" height="{height}" as="geometry" />
        </mxCell>'''
    
    for i, (f_type, f_name) in enumerate(fields):
        row_id = gen_id()
        label_id = gen_id()
        val_id = gen_id()
        bottom = 1 if i == 0 else 0
        xml += f'''
        <mxCell id="{row_id}" value="" style="shape=tableRow;horizontal=0;startSize=0;swimlaneHead=0;swimlaneBody=0;fillColor=none;collapsible=0;dropTarget=0;points=[[0,0.5],[1,0.5]];portConstraint=eastwest;top=0;left=0;right=0;bottom={bottom};" vertex="1" parent="{table_id}">
          <mxGeometry y="{30 + i*30}" width="{width}" height="30" as="geometry" />
        </mxCell>
        <mxCell id="{label_id}" value="{f_type}" style="shape=partialRectangle;connectable=0;fillColor=none;top=0;left=0;bottom=0;right=0;fontStyle=1;overflow=hidden;whiteSpace=wrap;html=1;" vertex="1" parent="{row_id}">
          <mxGeometry width="30" height="30" as="geometry">
            <mxRectangle width="30" height="30" as="alternateBounds" />
          </mxGeometry>
        </mxCell>
        <mxCell id="{val_id}" value="{f_name}" style="shape=partialRectangle;connectable=0;fillColor=none;top=0;left=0;bottom=0;right=0;align=left;spacingLeft=6;{"fontStyle=5;" if "PK" in f_type or "FK" in f_type else ""}overflow=hidden;whiteSpace=wrap;html=1;" vertex="1" parent="{row_id}">
          <mxGeometry x="30" width="{width-30}" height="30" as="geometry">
            <mxRectangle width="{width-30}" height="30" as="alternateBounds" />
          </mxGeometry>
        </mxCell>'''
    return xml, table_id

tables = [
    ('PUSH TOKEN', [('PK', 'ID'), ('FK', 'UserID'), ('', 'Token'), ('', 'Platform'), ('', 'CreatedAt'), ('', 'UpdatedAt')], 450, -700),
    ('GROUP INVITATION', [('PK', 'Code'), ('FK', 'GroupID'), ('FK', 'CreatedBy'), ('', 'ExpiresAt'), ('', 'CreatedAt')], 200, -600),
    ('MESSAGE THREAD', [('PK', 'ID'), ('FK', 'UserA_ID'), ('FK', 'UserB_ID'), ('', 'CreatedAt'), ('', 'UpdatedAt')], 450, -400),
    ('MESSAGE', [('PK', 'ID'), ('FK', 'ThreadID'), ('FK', 'SenderID'), ('', 'Content'), ('', 'IsRead'), ('', 'CreatedAt')], 650, -400),
    ('NOTIFICATION', [('PK', 'ID'), ('FK', 'UserID'), ('', 'Title'), ('', 'Message'), ('', 'IsRead'), ('', 'CreatedAt'), ('', 'UpdatedAt')], 650, -700),
    ('OTP CODE', [('PK', 'ID'), ('', 'Identifier'), ('', 'Destination'), ('', 'Code'), ('', 'Purpose'), ('', 'Payload'), ('', 'ExpiresAt'), ('', 'IsUsed'), ('', 'CreatedAt')], 850, -700, 200),
    ('SUBSCRIPTION PLAN', [('PK', 'ID'), ('', 'Name'), ('', 'Description'), ('', 'Price'), ('', 'Currency'), ('', 'DurationDays'), ('', 'CreatedAt'), ('', 'UpdatedAt')], 450, -100, 200),
    ('USER SUBSCRIPTION', [('PK', 'ID'), ('FK', 'UserID'), ('FK', 'PlanID'), ('', 'Status'), ('', 'ReferenceID'), ('', 'StartDate'), ('', 'EndDate'), ('', 'CreatedAt'), ('', 'UpdatedAt')], 700, -100, 200),
]

for name, fields, x, y, *extra_w in tables:
    width = extra_w[0] if extra_w else 180
    xml, tid = create_table(name, fields, x, y, width)
    print(f'<!-- {name} -->')
    print(xml)
