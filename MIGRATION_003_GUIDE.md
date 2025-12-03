# Migration 003: Major Redesign - Implementation Guide

## Overview

This migration implements a complete architectural redesign that separates student management from parent registration and adds support for teachers, attendance tracking, and test results.

## Key Changes

### 1. **New Architecture**
- **Before**: Parents registered with child information (name + class) embedded
- **After**: Students are separate entities managed by admins/teachers and linked to parents

### 2. **Multi-Child Support**
- Parents can now have up to **5 children** linked to their account
- **IMPORTANT:** Each student can only be linked to ONE parent (not multiple parents)
- Parents select which child to view information for

### 3. **New Roles**
- **Teachers**: Can manage students, post announcements, track attendance
- **Students**: Managed independently with their own records
- **Parents**: Register with just phone + language, children linked by admin

### 4. **New Features**
- Student management by admins/teachers
- Test results tracking
- Attendance tracking
- Multi-class announcement targeting

## Database Schema Changes

### New Tables

1. **teachers** - Teacher accounts with phone-based authentication
2. **students** - Student records (independent of parents)
3. **parent_students** - Junction table linking parents to students (max 5 per parent, ONE parent per student)
4. **teacher_classes** - Junction table for teacher-class assignments
5. **announcement_classes** - Junction table for multi-class announcements
6. **test_results** - Student test scores
7. **attendance** - Daily attendance records

### Modified Tables

1. **users** (parents)
   - Removed: `child_name`, `child_class`
   - Added: `current_selected_student_id` (which child is currently selected)

2. **announcements**
   - Added: `posted_by_teacher_id` (teachers can now post)
   - Added: Multi-class targeting via `announcement_classes` junction table

### New Views

1. **v_students_with_class** - Students with class information
2. **v_parent_children** - Parent-child relationships with full details
3. **v_test_results_detailed** - Test results with student and class info
4. **v_attendance_detailed** - Attendance with student and class info

## New Bot Commands

### For Admins

1. `/add_student` - Add a new student to the system
   ```
   Format:
   Student Name Surname
   Class

   Example:
   Aziz Karimov
   5-A
   ```

2. `/link_student` - Link an existing student to a parent
   ```
   Format:
   Parent Phone (+998XXXXXXXXX)
   Student ID

   Example:
   +998901234567
   15
   ```

3. `/list_students` - List all students with their IDs
   - Shows: ID, Name, Class, Active status

4. `/view_parent_children` - View children linked to a parent
   ```
   Format:
   Enter parent phone: +998XXXXXXXXX
   ```
   - Shows: All children with current selection marker 🎯

### For Parents

1. `/my_children` - View and switch between children
   - Shows inline buttons for each child
   - Current selection marked with ✅
   - Tap to switch active child

2. **Button: "👨‍👩‍👧‍👦 Mening farzandlarim / Мои дети"**
   - Added to main menu for parents with multiple children
   - Quick access to child selection

## Registration Flow Changes

### Old Flow (Deprecated)
1. Select language
2. Share phone number
3. Enter child name
4. Select child class
5. Complete registration

### New Flow (Current)
1. Select language
2. Share phone number
3. **Registration complete!**
4. Admin links students separately

### Why This Change?
- **More accurate data**: Admin controls student records
- **Multi-child support**: Parents can have multiple children
- **Flexibility**: Students can be added/removed without affecting parent account
- **Data integrity**: One source of truth for student information

## Implementation Steps Completed

✅ Step 1: Student Management Handlers
- Created `internal/handlers/student_management.go`
- Handlers for add, link, list, and view operations

✅ Step 2: Student Selection for Parents
- Created `internal/handlers/student_selection.go`
- Parents can view and switch between children
- Inline keyboard with visual indicators

✅ Step 3: Router Updates
- Added new command routes
- Added new state handlers
- Added callback handlers for child selection

✅ Step 4: Database Migration
- Migration 003 applied successfully
- All tables and views created
- Triggers and constraints in place

✅ Step 5: Repository Methods
- Added `GetByPhone` alias to UserRepository
- Student repository already had all needed methods
- Junction table support via existing methods

## Usage Examples

### Example 1: Admin adds a student
```
Admin: /add_student
Bot: Please send student info in format:
     First Last Name
     Class

Admin: Aziz Karimov
       5-A

Bot: ✅ Student successfully added!
     ID: 15
     Name: Aziz Karimov
     Class: 5-A
```

### Example 2: Admin links student to parent
```
Admin: /link_student
Bot: Please send in format:
     Parent Phone
     Student ID

Admin: +998901234567
       15

Bot: ✅ Successfully linked!
     Parent: +998901234567
     Student: Aziz Karimov (ID: 15)
     Class: 5-A

[Parent receives notification]
Parent: 👨‍👩‍👧‍👦 New child linked!
        Name: Aziz Karimov
        Class: 5-A
```

### Example 3: Parent with multiple children
```
Parent: /my_children
Bot: 👨‍👩‍👧‍👦 Your Children

     [Button: Aziz Karimov ✅ - 5-A]
     [Button: Laylo Karimova - 3-B]

Parent: *taps Laylo button*
Bot: ✅ Child selected
     Name: Laylo Karimova
     Class: 3-B

     Now all information (timetable, grades, attendance)
     will be shown for this child.
```

## Data Migration Notes

### Existing Data
- Old users have `child_name` and `child_class` in state data (deprecated but kept for compatibility)
- **Action required**: Admins should:
  1. Create student records for all existing parent-child pairs
  2. Link students to parents using `/link_student`
  3. This ensures data continuity

### Backward Compatibility
- Old state fields kept in `StateData` model with deprecation comments
- Old handlers (`HandleChildName`, `HandleChildClass`) redirect to new flow
- No breaking changes for active users

## Testing Checklist

- [x] Migration applies without errors
- [x] All tables created with correct schema
- [x] All views created successfully
- [x] Foreign keys and constraints work
- [x] Commands route correctly
- [x] Handlers compile without errors
- [ ] End-to-end test: Admin adds student
- [ ] End-to-end test: Admin links student to parent
- [ ] End-to-end test: Parent selects between children
- [ ] End-to-end test: Parent views timetable for selected child
- [ ] Test max constraints (5 children per parent, 3 admins)

## Security Considerations

1. **Admin Verification**: All student management commands verify admin status
2. **Parent Verification**: Child selection verifies student belongs to parent
3. **Data Constraints**: Database enforces max limits (4 children, 3 admins)
4. **Phone Validation**: Uzbek phone format enforced (+998XXXXXXXXX)

## Performance Optimizations

1. **Indexed Queries**: All foreign keys indexed
2. **Views**: Pre-joined data for common queries
3. **Pagination**: All list operations support limit/offset
4. **Junction Tables**: Efficient many-to-many relationships

## Future Enhancements

### Planned
- [ ] Bulk student import (CSV/Excel)
- [ ] Parent can request child linkage (pending admin approval)
- [ ] Teacher can view their class students
- [ ] Teacher can add test results and attendance
- [ ] Parent notifications for new grades/attendance
- [ ] Analytics dashboard for admins

### Under Consideration
- [ ] Multiple parents per student (divorced parents)
- [ ] Student transfer between classes
- [ ] Historical grade tracking
- [ ] Attendance reports
- [ ] Parent-teacher messaging

## Troubleshooting

### Issue: "Student not found"
- **Cause**: Student ID doesn't exist
- **Solution**: Use `/list_students` to get correct IDs

### Issue: "Maximum 5 children allowed"
- **Cause**: Parent already has 4 children linked
- **Solution**: Remove a child before adding new one (admin operation)

### Issue: "Parent not found"
- **Cause**: Phone number not registered
- **Solution**: Parent must register first with `/start`

### Issue: "This student already linked"
- **Cause**: Attempting to link same student twice
- **Solution**: Check `/view_parent_children` for existing links

## Support

For issues or questions:
1. Check this guide first
2. Review the migration SQL file: `internal/database/migrations/003_major_redesign.sql`
3. Check handler implementations in `internal/handlers/student_*.go`
4. File an issue on GitHub

## Rollback Procedure

If rollback is needed:
1. Backup database: `cp parent_bot.db parent_bot.db.backup`
2. Restore previous schema (requires custom rollback script)
3. **Note**: Rollback will lose all new data (students, teachers, test results, attendance)

**Recommendation**: Test thoroughly in development before production deployment.
