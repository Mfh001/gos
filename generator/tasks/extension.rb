def convert_define(type)
    case type.to_s
    when 'integer'
        'int'
    when 'boolean'
        'bool'
    when 'float'
        'float32'
    else
        type.to_s
    end
end

def type_formatter(type)
  type = convert_define(type)
  case type
  when 'int'
    '%d'
  when 'float'
    '%d'
  when 'bool'
    '%t'
  when 'string'
    "'%s'"
  else
    raise "invalid type: #{type}"
  end
end

def get_struct_name(table_name)
    table_name.singularize.camelize
end

class String
  def is_number?
    true if Float(self) rescue false
  end
  def uncapitalize
    self[0, 1].downcase + self[1..-1]
  end
end

def gen_struct(table_name, field_names, field_types, db_tag = false)
    struct_content = "type #{get_struct_name(table_name)} struct{\n"
    field_names.each_with_index do |field_name, index|
      field_type = field_types[index]
      if db_tag
        struct_content << %Q{    #{field_name.camelize} #{convert_define(field_type)} `db:"#{c.name}"`\n}
      else
        struct_content << %Q{    #{field_name.camelize} #{convert_define(field_type)}\n}
      end
    end
    struct_content << "}"
    return struct_content
end
