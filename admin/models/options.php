<?php
/**
 * 选项表
 */

class TB_Options
{
    /**
     * The keyed value of $name
     * 
     * @param string $name the key name
     * @param mixed  $def  the default value if key does not exist
     * 
     * @return the keyed value of $name
     */
    public function get($name, $def='')
    {
        global $tbdb;
        $sql = 'SELECT value FROM options WHERE name=? LIMIT 1';
        if ($stmt = $tbdb->prepare($sql)) {
            $stmt->bind_param('s', $name);
            $stmt->execute();
            $stmt->bind_result($value);
            $stmt->fetch();
            $stmt->close();

            return $value ? $value : $def;
        }

        return $def;
    }

    /**
     * Sets key's value
     * 
     * @param string $name  the key name
     * @param string $value the key's value
     * 
     * @return boolean
     */
    public function set($name, $value)
    {
        global $tbdb;
        if ($this->has($name)) {
            $sql = 'UPDATE options SET value=? WHERE name=? LIMIT 1';
            if ($stmt = $tbdb->prepare($sql)) {
                $stmt->bind_param('ss', $value, $name);
                return $stmt->execute();
            }
        } else {
            $sql = 'INSERT INTO options (name,value) VALUES (?,?)';
            if ($stmt = $tbdb->prepare($sql)) {
                $stmt->bind_param('ss', $name, $value);
                return $stmt->execute();
            }
        }

        return false;
    }

    /**
     * Tests if the value specified by key $name exists
     * 
     * @param string $name the key name
     * 
     * @return boolean
     */
    public function has($name)
    {
        global $tbdb;
        $sql = 'SELECT name FROM options WHERE name=? LIMIT 1';
        if ($stmt = $tbdb->prepare($sql)) {
            $stmt->bind_param('s', $name);
            $stmt->execute();
            $ret = $stmt->get_result();
            return $ret!==false && $ret->num_rows>0;
        }

        return false;
    }

    /**
     * Deletes value specified by key name
     * 
     * @param string $name the key name
     * 
     * @return boolean
     */
    public function del($name)
    {
        global $tbdb;
        $sql = 'DELETE FROM options WHERE name=? LIMIT 1';
        if ($stmt = $tbdb->prepare($sql)) {
            $stmt->bind_param('s', $name);
            return $stmt->execute();
        }

        return false;
    }
}
